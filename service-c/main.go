package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	sarama "github.com/Shopify/sarama"
	oauth "github.com/damiannolan/sasl/oauthbearer"
	chi "github.com/go-chi/chi/v5"
)

func main() {
	clientID := "service-c"
	clientSecret := "service-c-secret"
	tokenURL := "http://keycloak:8080/auth/realms/meetup/protocol/openid-connect/token"

	r := chi.NewRouter()
	// r.Use(middleware.Logger)

	tokenProvider := oauth.NewTokenProvider(clientID, clientSecret, tokenURL)

	//Invoke HTTP endpoint
	r.Get("/send2http", func(w http.ResponseWriter, r *http.Request) {
		serviceDendpoint := "http://localhost:8000/v1/users"

		accessToken, err := tokenProvider.Token()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ERROR: " + err.Error()))
			return
		}

		client := &http.Client{}

		req, err := http.NewRequest("GET", serviceDendpoint, nil)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ERROR: " + err.Error()))
			return
		}

		req.Header.Add("Authorization", "Bearer "+accessToken.Token)

		res, err := client.Do(req)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ERROR: " + err.Error()))
			return
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ERROR: " + err.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("RECEIVED: " + string(body)))
		}
	})

	// Send to Kafka topic
	r.Get("/send2kafka", func(w http.ResponseWriter, r *http.Request) {
		kafkaBrokers := strings.Split("localhost:9092", ",")
		kafkaTopic := "a_messages"

		requestId := rand.Intn(9000) + 999
		msg := "test message with id=" + strconv.Itoa(requestId) + " from service-c"

		_, _, err := kafkaProducer(kafkaBrokers, tokenProvider, kafkaTopic, msg)
		w.Header().Add("Content-Type", "text/plain")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ERROR: " + err.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("SENT: " + msg))
		}
	})
	log.Fatal(http.ListenAndServe(":8383", r))
}

func kafkaProducer(brokerList []string, tokenProvider sarama.AccessTokenProvider, topic string, msg string) (int32, int64, error) {

	config := sarama.NewConfig()

	config.Version = sarama.MaxVersion

	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
	config.Net.SASL.TokenProvider = tokenProvider

	syncProducer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create producer: %v", err)
	}
	defer syncProducer.Close()

	partition, offset, err := syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send message to %s: %v", topic, err)
	}

	log.Printf("wrote [%s] message at partition: %d, offset: %d\n", msg, partition, offset)
	return partition, offset, nil
}
