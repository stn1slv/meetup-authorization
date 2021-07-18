package main

import (
	"log"
	"strconv"
	"strings"

	"math/rand"
	"net/http"

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
	clientID := "service-c"
	clientSecret := "service-c-secret"
	tokenURL := "http://keycloak:8080/auth/realms/meetup/protocol/openid-connect/token"
	splitBrokers := strings.Split("localhost:9092", ",")
	topic := "a_messages"

	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	config.Version = sarama.MaxVersion

	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
	config.Net.SASL.TokenProvider = oauth.NewTokenProvider(clientID, clientSecret, tokenURL)

	syncProducer, err := sarama.NewSyncProducer(splitBrokers, config)
	if err != nil {
		log.Println("failed to create producer: ", err)
		return 0, 0, err
	}
	defer syncProducer.Close()
	partition, offset, err := syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	})
	if err != nil {
		log.Printf("failed to send message to %s: %e\n", topic, err)
		return 0, 0, err
	}
	log.Printf("wrote message at partition: %d, offset: %d\n", partition, offset)
	return partition, offset, nil
}
