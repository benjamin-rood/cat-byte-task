package main

import (
	"net/http"

	"twitch_chat_analysis/internal/messaging"
	"twitch_chat_analysis/internal/rabbitmq"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	rabbitMQ, err := rabbitmq.NewRabbitMQ(messaging.RabbitMQURL)
	if err != nil {
		panic(err)
	}
	defer rabbitMQ.Close()

	r.POST("/message", func(c *gin.Context) {
		var msg messaging.Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := rabbitMQ.PublishMessage(msg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
			return
		}

		c.Status(http.StatusOK)
	})

	r.Run(":8080")
}
