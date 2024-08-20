package main

import (
	"net/http"
	"sort"

	"twitch_chat_analysis/internal/messaging"
	"twitch_chat_analysis/internal/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	redisClient, err := redis.NewRedisClient("redis:6379")
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	r.GET("/message/list", func(c *gin.Context) {
		sender := c.Query("sender")
		receiver := c.Query("receiver")

		if sender == "" || receiver == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both sender and receiver are required"})
			return
		}

		messages, err := redisClient.GetMessages(sender, receiver)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
			return
		}

		// Sort messages in chronological descending order
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Timestamp > messages[j].Timestamp
		})

		result := make([]messaging.Message, len(messages))
		for i, msg := range messages {
			result[i] = messaging.Message{
				Sender:   sender,
				Receiver: receiver,
				Message:  msg.Content,
			}
		}

		c.JSON(http.StatusOK, result)
	})

	r.Run(":8081")
}
