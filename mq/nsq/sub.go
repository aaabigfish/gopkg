package nsq

import (
	"fmt"

	"github.com/nsqio/go-nsq"
)

type HandlerFunc func(*NsqMessage) error

type Handler interface {
	HandleMessage(message *NsqMessage) error
}

// HandleMessage implements the Handler interface
func (h HandlerFunc) HandleMessage(m *NsqMessage) error {
	return h(m)
}

type Reader interface {
	GetConsumer() *nsq.Consumer
	AddHandler(HandlerFunc)
	AddHandlers(int, HandlerFunc)
	Run() error
	Close()
}

type reader struct {
	topic    string
	channel  string
	addrs    []string
	c        *NsqConfig
	consumer *nsq.Consumer
}

func NewConsumer(topic, channel string, addrs []string, c ...*NsqConfig) Reader {
	cfg := &NsqConfig{}
	if len(c) > 0 {
		cfg = c[0]
	} else {
		cfg = NewNsqConfig()
	}

	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		panic(fmt.Sprintf("new nsq consumer failed err(%v)", err))
	}

	return &reader{
		topic:    topic,
		channel:  channel,
		addrs:    addrs,
		c:        cfg,
		consumer: consumer,
	}
}

func (r *reader) GetConsumer() *nsq.Consumer {
	return r.consumer
}

func (r *reader) AddHandler(hookFunc HandlerFunc) {
	r.consumer.AddHandler(HandlerFunc(hookFunc))
}

// 并发处理消息，n是并发数量
func (r *reader) AddHandlers(n int, hookFunc HandlerFunc) {
	r.consumer.AddConcurrentHandlers(HandlerFunc(hookFunc), n)
}

func (r *reader) Run() error {
	return r.consumer.ConnectToNSQLookupds(r.addrs)
}

func (r *reader) Close() {
	if r.consumer != nil {
		r.consumer.Stop()
		return
	}
}
