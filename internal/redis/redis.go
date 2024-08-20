package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	connectionFailures := 0

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	// Wait for Redis to be ready
	for {
		_, err := client.Ping(context.Background()).Result()
		if err == nil {
			break
		}
		connectionFailures++
		if connectionFailures >= 5 {
			return nil, fmt.Errorf("failed to connect to Redis after multiple attempts")
		}
		time.Sleep(2 * time.Second)
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) SaveMessage(sender, receiver, message string) error {
	key := fmt.Sprintf("messages:%s:%s", sender, receiver)
	member := fmt.Sprintf("%d:%s", time.Now().UnixNano(), message)

	err := r.client.ZAdd(r.client.Context(), key, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: member,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}

	return nil
}

func (r *RedisClient) GetMessages(sender, receiver string) ([]struct {
	Timestamp int64
	Content   string
}, error) {
	key := fmt.Sprintf("messages:%s:%s", sender, receiver)

	result, err := r.client.ZRevRange(r.client.Context(), key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %v", err)
	}

	messages := make([]struct {
		Timestamp int64
		Content   string
	}, len(result))

	for i, msg := range result {
		parts := strings.SplitN(msg, ":", 2)
		if len(parts) != 2 {
			continue
		}
		timestamp, _ := strconv.ParseInt(parts[0], 10, 64)
		messages[i] = struct {
			Timestamp int64
			Content   string
		}{
			Timestamp: timestamp,
			Content:   parts[1],
		}
	}

	return messages, nil
}
