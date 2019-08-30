package chremoas_redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type Client struct {
	Client *redis.Client
	Prefix string
}

var Nil = redis.Nil

func Init(prefix string) *Client {
	var c *redis.Client

	if viper.GetStringSlice("redis.sentinels") != nil {
		c = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    viper.GetString("redis.sentinelMasterName"),
			SentinelAddrs: viper.GetStringSlice("redis.sentinels"),
			Password: viper.GetString("redis.password"), // no password set
			DB:       viper.GetInt("redis.database"),    // use default DB
		})
	} else {
		c = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d",
				viper.GetString("redis.host"),
				viper.GetInt("redis.port")),
			Password: viper.GetString("redis.password"), // no password set
			DB:       viper.GetInt("redis.database"),    // use default DB
		})
	}

	return &Client{Client: c, Prefix: prefix}
}

func (r Client) KeyName(key string) string {
	return fmt.Sprintf("%s:%s", r.Prefix, key)
}
