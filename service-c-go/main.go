package main

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	oauth "github.com/damiannolan/sasl/oauthbearer"
)

func main() {
	clientID := "service-c"
	clientSecret := "service-c-secret"
	tokenURL := "http://keycloak:8080/auth/realms/meetup/protocol/openid-connect/token"
	splitBrokers := strings.Split("localhost:9092", ",")

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
		fmt.Println("failed to create producer: ", err)
	}
	partition, offset, err := syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: "a_messages",
		Value: sarama.StringEncoder("test_message"),
	})
	if err != nil {
		fmt.Printf("failed to send message to a_messages: %e", err)
	}
	fmt.Printf("wrote message at partition: %d, offset: %d", partition, offset)
	_ = syncProducer.Close()
}
