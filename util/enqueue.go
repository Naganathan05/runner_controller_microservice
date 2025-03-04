package util

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

func EnqueueRunRequest(ctx context.Context, runID string, fileName string, extension string) error {

	// Message represents the structure of our message
	type Message struct {
		RunId	string	`json:"runId"`
		FileName	string	`json:"fileName"`
		Extension	string	`json:"extension"`
		Timestamp	time.Time	`json:"timestamp"`
	}
	
	// Get RabbitMQ connection string from environment variable or use default
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	// Connect to RabbitMQ server
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare a queue
	queueName := "task_queue"
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}


	// Create a new message
	msg := Message{
		RunId: runID,
		FileName: fileName,
		Extension: extension,
		Timestamp: time.Now(),
	}

	// Convert message to JSON
	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
	}

	// Publish message
	err = ch.PublishWithContext(
		ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, 
			Body: body,
		},
	)
			
	if err != nil {
		log.Printf("Failed to publish a message: %v", err)
	}
	
	log.Printf("Published message: %s", msg.RunId)
		

	return nil
}