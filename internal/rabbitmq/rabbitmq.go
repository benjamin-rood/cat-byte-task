package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"twitch_chat_analysis/internal/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	q       amqp.Queue
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	var (
		conn *amqp.Connection
		err  error
	)

	connectionFailures := 0

	// Wait for RabbitMQ to be ready
	for {
		conn, err = amqp.Dial("amqp://user:password@rabbitmq:5672/")
		if err == nil {
			break
		}
		connectionFailures++
		if connectionFailures >= 5 {
			return nil, fmt.Errorf("failed to connect to RabbitMQ after multiple attempts")
		}
		time.Sleep(2 * time.Second)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		messaging.RabbitQueueName, // name
		false,                     // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		q:       q,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}
}

func (r *RabbitMQ) PublishMessage(msg messaging.Message) error {

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := r.channel.Publish(
		"",       // exchange
		r.q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}); err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	return nil
}

func (r *RabbitMQ) Consumer() (<-chan amqp.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.q.Name, // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}

	return msgs, nil
}
