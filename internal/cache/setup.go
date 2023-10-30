package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
)

func InitCache() {

	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "mysecretpassword",
		DB:       0,
	})
}

func Save(key string, content string) error {

	ctx := context.Background()

	err := client.Set(ctx, key, content, 0).Err()
	if err != nil {
		log.Fatal(err)
		return err
	}

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println("Read from redis:", val)

	return nil
}
