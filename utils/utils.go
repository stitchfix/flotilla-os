package utils

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/stitchfix/flotilla-os/config"
	"strings"
)

// StringSliceContains checks is a string slice contains a particular string.
func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func SetupRedisClient(c config.Config) (*redis.Client, error) {
	if !c.IsSet("redis_address") {
		return nil, fmt.Errorf("redis_address not configured")
	}

	redisAddress := strings.TrimPrefix(c.GetString("redis_address"), "redis://")
	redisDB := c.GetInt("redis_db")

	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
		DB:   redisDB,
	})

	_, err := client.Ping().Result()

	return client, err
}
