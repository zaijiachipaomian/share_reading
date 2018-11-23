package utils

import (
	"fmt"
	"github.com/go-redis/redis"
	"sync"
)

var (
	cli *redis.Client
	one sync.Once
)

func init() {

	one.Do(func() {
		cli = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 5,
		})
		_, err := cli.Ping().Result()
		if err != nil {
			panic("cli.Ping .Result " + err.Error())
		} else {
			fmt.Println("redis success ")
		}
	})
}

func GetClient() (client *redis.Client) {
	client = cli
	return client
}
