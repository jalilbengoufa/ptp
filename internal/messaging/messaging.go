package messaging

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jalilbengoufa/ptp/internal"
	"github.com/jalilbengoufa/ptp/internal/database"
	"github.com/jalilbengoufa/ptp/internal/database/postgres"
)

var (
	Producer *kafka.Producer
	Consumer *kafka.Consumer
)

func InitKafkaProducer() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "ptp-client",
		"acks":              "all"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)

	}
	Producer = p
	defer p.Close()

	// Delivery report handler for produced messages
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
			} else {
				fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
			}
		}
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
}

func InitKafkaConsumer() {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "ptp-group",
		"auto.offset.reset": "smallest"})

	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)

	}

	Consumer = consumer
	consumer.SubscribeTopics([]string{"fileEdit"}, nil)
	if err != nil {
		fmt.Printf("Failed to SubscribeTopics: %s\n", err)

	}

	// A signal handler or similar could be used to set this to false to break the loop.
	defer consumer.Close()

	for {
		msg, err := consumer.ReadMessage(time.Second)
		if err == nil {
			fmt.Printf("Message on %s: %s =========\n", msg.TopicPartition, string(msg.Value))

			// Save to postgres
			file := database.File{ID: "1", Name: internal.Filename, Content: string(msg.Value)}
			postgres.DbInstancePostgres.Save(&file)

			// Read from postgres
			var fileTextPostgres database.File
			postgres.DbInstancePostgres.First(&fileTextPostgres, file.ID)
			fmt.Println("Read from postgres:", fileTextPostgres.Content)
		} else {
			// The client will automatically try to recover from all errors.
			// Timeout is not considered an error because it is raised by
			// ReadMessage in absence of messages.
			// fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

}
