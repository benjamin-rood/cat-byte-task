package main

import (
	"encoding/json"
	"log"

	"twitch_chat_analysis/internal/messaging"
	"twitch_chat_analysis/internal/rabbitmq"
	"twitch_chat_analysis/internal/redis"
)

func main() {

	rabbitMQ, err := rabbitmq.NewRabbitMQ(messaging.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	redisClient, err := redis.NewRedisClient(messaging.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	msgs, err := rabbitMQ.Consumer()
	if err != nil {
		log.Fatalf("Failed to get the consumer: %v", err)
	}

	for msg := range msgs {
		var message messaging.Message
		err := json.Unmarshal(msg.Body, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		err = redisClient.SaveMessage(message.Sender, message.Receiver, message.Message)
		if err != nil {
			log.Printf("Error saving message to Redis: %v", err)
		}
	}
}
