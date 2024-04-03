package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
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

func setOptions(opts *gredis.UniversalOptions, dsn string) {
	url, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}

	args := url.Query()

	rv := reflect.ValueOf(opts).Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if !f.CanInterface() {
			continue
		}
		name := rt.Field(i).Name
		arg := args.Get(name)
		if arg == "" {
			continue
		}
		switch f.Interface().(type) {
		case time.Duration:
			v, err := time.ParseDuration(arg)
			if err != nil {
				panic(fmt.Sprintf("%s=%s, err:%v", name, arg, err))
			}
			f.Set(reflect.ValueOf(v))
		case int:
			v, err := strconv.Atoi(arg)
			if err != nil {
				panic(fmt.Sprintf("%s=%s, err:%v", name, arg, err))
			}
			f.SetInt(int64(v))
		case bool:
			v, err := strconv.ParseBool(arg)
			if err != nil {
				panic(fmt.Sprintf("%s=%s, err:%v", name, arg, err))
			}
			f.SetBool(v)
		case string:
			f.SetString(arg)
		}
	}

	opts.Addrs = []string{url.Host}
	opts.Username = url.User.Username()
	if p, ok := url.User.Password(); ok {
		opts.Password = p
	}

	// 获取database
	f := strings.FieldsFunc(url.Path, func(r rune) bool {
		return r == '/'
	})
	switch len(f) {
	case 0:
		opts.DB = 0
	case 1:
		var err error
		if opts.DB, err = strconv.Atoi(f[0]); err != nil {
			panic(fmt.Sprintf("redis: invalid database number: %q", f[0]))
		}
	default:
		panic(fmt.Sprintf("redis: invalid URL path: %s", url.Path))
	}
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
