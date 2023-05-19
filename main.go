package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	kafkaHost := os.Getenv("KAFKA_HOST")
	groupId := os.Getenv("GROUP")
	offset := os.Getenv("OFFSET")

	if kafkaHost == "" {
		kafkaHost = "localhost:9092"
	}

	if groupId == "" {
		groupId = "golang-kafka-group"
	}

	if offset == "" {
		offset = kafka.OffsetBeginning.String()
	}

	// Set up Kafka consumer configuration
	consumerConfig := &kafka.ConfigMap{
		"bootstrap.servers": kafkaHost,
		"group.id":          groupId,
		"auto.offset.reset": offset,
	}

	// Create new consumer
	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Subscribe to Kafka topic
	err = consumer.SubscribeTopics([]string{"input-topic"}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %v", err)
	}

	// Set up Kafka producer configuration
	producerConfig := &kafka.ConfigMap{
		"bootstrap.servers": kafkaHost,
	}

	// Create new producer
	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Set up signal handling to gracefully stop the consumer
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Consume messages from Kafka topic
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case ev := <-consumer.Events():
				switch e := ev.(type) {
				case *kafka.Message:
					// Process incoming message
					fmt.Printf("Received message: %s\n", string(e.Value))

					// Create new message to send back to Kafka
					producedMsg := &kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &[]string{os.Getenv("TOPIC")}[0], Partition: e.TopicPartition.Partition},
						Value:          e.Value,
					}

					// Send message back to Kafka
					err := producer.Produce(producedMsg, nil)
					if err != nil {
						log.Printf("Failed to send message: %v", err)
					}

				case kafka.Error:
					log.Printf("Consumer error: %v", e)
				}

			case <-signals:
				log.Println("Stopping consumer...")
				return
			}
		}
	}()

	// Wait for the consumer to finish
	wg.Wait()

	log.Println("Consumer stopped.")
}
