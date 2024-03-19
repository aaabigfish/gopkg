package redis

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	gredis "github.com/redis/go-redis/v9"
)

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
