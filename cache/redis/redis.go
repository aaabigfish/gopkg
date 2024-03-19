package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gredis "github.com/redis/go-redis/v9"
)

// Client redis 客户端
type Client struct {
	gredis.UniversalClient
}

// Get 获取缓存实例
func NewClient(dsn string) *Client {
	opts := &gredis.UniversalOptions{}

	setOptions(opts, dsn)

	rdb := gredis.NewUniversalClient(opts)

	db := &Client{rdb}

	return db

}

func (c *Client) SaveStruct(ctx context.Context, key string, data interface{}, second int) error {
	bData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("save key(%s) json.Marshal(%v) err(%v)", key, data, err)
	}
	err = c.Set(ctx, key, string(bData), time.Duration(second)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("save key(%s) fail:%s", key, err.Error())
	}
	return nil
}

func (c *Client) GetStruct(ctx context.Context, key string, data interface{}) error {
	valStr, err := c.Get(ctx, key).Result()
	if gredis.Nil == err {
		return fmt.Errorf("key(%s) not found", key)
	}
	if err != nil {
		return fmt.Errorf("get key(%s) fail:%s", key, err.Error())
	}
	return json.Unmarshal([]byte(valStr), &data)
}
