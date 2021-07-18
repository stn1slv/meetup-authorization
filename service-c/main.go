package main

import (
	"fmt"
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
	r := chi.NewRouter()
	// r.Use(middleware.Logger)

	r.Get("/send2http", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("sent to HTTP"))
	})

	r.Get("/send2kafka", func(w http.ResponseWriter, r *http.Request) {
		requestId := rand.Intn(9000) + 999
		msg := "test message with id=" + strconv.Itoa(requestId) + " from service-c"

		_, _, err := send2kafka(msg)
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

func send2kafka(msg string) (int32, int64, error) {
	splitBrokers := strings.Split("localhost:9092", ",")
	topic := "a_messages"

	clientID := "service-c"
	clientSecret := "service-c-secret"
	tokenURL := "http://keycloak:8080/auth/realms/meetup/protocol/openid-connect/token"

	tokenProvider := oauth.NewTokenProvider(clientID, clientSecret, tokenURL)
	accessToken, err := tokenProvider.Token()
	if err != nil {
		return 0, 0, err
	}
	log.Println(accessToken.Token)

	return kafkaProducer(splitBrokers, tokenProvider, topic, msg)
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
		// log.Println("failed to create producer: ", err)
		return 0, 0, fmt.Errorf("failed to create producer: %v", err)
	}
	defer syncProducer.Close()

	partition, offset, err := syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	})
	if err != nil {
		// log.Printf("failed to send message to %s: %e\n", topic, err)
		return 0, 0, fmt.Errorf("failed to send message to %s: %v", topic, err)
	}
	log.Printf("wrote [%s] message at partition: %d, offset: %d\n", msg, partition, offset)
	return partition, offset, nil
}
