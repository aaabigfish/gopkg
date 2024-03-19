package nsq

import (
	"github.com/nsqio/go-nsq"
)

type NsqConfig = nsq.Config
type NsqMessage = nsq.Message

type Config struct {
	topic     string      // 默认推送的topic，如果设置了发送消息就不用指定topic
	NsqConfig *nsq.Config // nsq的配置
}

func NewConfig(topics ...string) *Config {
	topic := ""
	if len(topics) > 0 {
		topic = topics[0]
	}
	return &Config{
		topic:     topic,
		NsqConfig: nsq.NewConfig(),
	}
}

func NewNsqConfig() *nsq.Config {
	return nsq.NewConfig()
}
