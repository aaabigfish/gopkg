package nsq

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
)

type Producer = nsq.Producer

var NsqPublisher = &nsq.Producer{}

type Writer interface {
	Push(data interface{}, key ...[]byte) error
	PushTopic(topic string, data interface{}, key ...[]byte) error
	PushMessage(msg *Message) error
	Publish(topic string, body []byte, key ...[]byte) error
	MultiPublish(topic string, bodys [][]byte, key ...[]byte) error
	PublishDelay(topic string, t time.Duration, body []byte, key ...[]byte) error
	GetProducer() *Producer
	Close()
}

type producer struct {
	id        int       // 生产者id
	pub       *Producer // 生产者
	unPubFunc func()    // 解除函数
}

type writer struct {
	c         *Config
	topic     string
	addrs     []string
	mu        sync.RWMutex
	producers []*producer
	balancer  Balancer
}

// 创建生产者
func NewProducer(addr string, c ...*Config) Writer {
	cfg := &Config{}
	if len(c) > 0 {
		cfg = c[0]
	} else {
		cfg = NewConfig()
	}

	producer, err := nsq.NewProducer(addr, cfg.NsqConfig)
	if err != nil {
		panic(fmt.Sprintf("new nsq producer failed err(%v)", err))
	}

	w := &writer{
		addrs: []string{addr},
		c:     cfg,
		topic: cfg.topic,
	}
	w.addProducer(producer)

	return w
}

// 创建生产者链接池
// balancer 平衡器，在nsq集群上将消息路由到可用节点的逻辑，详见balancer.go
func NewProducerPool(addrs []string, balancer Balancer, c ...*Config) Writer {
	if len(addrs) == 0 {
		panic("addrs should not be empty")
	}

	cfg := &Config{}
	if len(c) > 0 {
		cfg = c[0]
	} else {
		cfg = NewConfig()
	}

	w := &writer{
		addrs: addrs,
		c:     cfg,
		topic: cfg.topic,
	}

	for _, addr := range addrs {
		p, err := nsq.NewProducer(addr, cfg.NsqConfig)
		if err != nil {
			panic(fmt.Sprintf("new nsq producer failed err(%v)", err))
		}
		w.addProducer(p)
	}

	if balancer != nil {
		w.balancer = balancer
	} else {
		w.balancer = &RoundRobin{}
	}

	return w
}

// 获取生产者实例
func (w *writer) GetProducer() *Producer {
	producer, _ := w.getProducer()
	if producer != nil {
		return nil
	}
	return producer.pub
}

// 获取平衡器，默认是循环所有节点
func (w *writer) getBalancer() Balancer {
	if w.balancer != nil {
		return w.balancer
	}
	return &RoundRobin{}
}

func (w *writer) addProducer(p *nsq.Producer) {
	w.mu.Lock()
	defer w.mu.Unlock()

	producer := &producer{
		pub: p,
	}

	producer.unPubFunc = func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		for i, pub := range w.producers {
			if pub == producer {
				if pub.pub != nil {
					pub.pub.Stop()
				}
				w.producers = append(w.producers[:i], w.producers[i+1:]...)
				return
			}
		}
	}

	w.producers = append(w.producers, producer)
}

func (w *writer) getProducer(key ...[]byte) (*producer, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	producersLen := len(w.producers)
	if producersLen <= 0 {
		return nil, errors.New("producers is empty")
	}
	if producersLen == 1 {
		if w.producers[0] == nil || w.producers[0].pub == nil {
			return nil, errors.New("producer is empty")
		}
		return w.producers[0], nil
	}

	var k []byte
	if len(key) > 0 {
		k = key[0]
	}
	producer := w.producers[w.getBalancer().Balance(k, loadCachedNodes(producersLen)...)]
	if producer == nil || producer.pub == nil {
		return nil, errors.New("producer is empty")
	}

	return producer, nil
}

// 发送消息（需要通过NewTopicProducer）
func (w *writer) Push(data interface{}, key ...[]byte) error {
	if w.topic == "" {
		return errors.New("topic is empty")
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return w.Publish(w.topic, payload, key...)
}

// 发送data到topic
func (w *writer) PushTopic(topic string, data interface{}, key ...[]byte) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return w.Publish(topic, payload, key...)
}

// 发送消息
func (w *writer) Publish(topic string, body []byte, key ...[]byte) error {
	producer, err := w.getProducer(key...)
	if err != nil {
		return err
	}

	if err := producer.pub.Publish(topic, body); err != nil {
		producer.unPubFunc()
		return err
	}
	return nil
}

// 批量发送消息
func (w *writer) MultiPublish(topic string, bodys [][]byte, key ...[]byte) error {
	producer, err := w.getProducer(key...)
	if err != nil {
		return err
	}

	if err := producer.pub.MultiPublish(topic, bodys); err != nil {
		producer.unPubFunc()
		return err
	}
	return nil
}

// 发送异常消息
func (w *writer) PublishDelay(topic string, t time.Duration, body []byte, key ...[]byte) error {
	producer, err := w.getProducer(key...)
	if err != nil {
		return err
	}
	if err := producer.pub.DeferredPublish(topic, t, body); err != nil {
		producer.unPubFunc()
		return err
	}
	return nil
}

func (w *writer) PushMessage(msg *Message) error {
	if w.topic == "" && msg.Topic == "" {
		return errors.New("topic is empty")
	}
	topic := w.topic
	if msg.Topic != "" {
		topic = msg.Topic
	}

	if msg.Time <= 0 {
		msg.Time = time.Now().UnixNano()
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return w.Publish(topic, payload, msg.Key)
}

func (w *writer) Close() {
	for _, p := range w.producers {
		p.unPubFunc()
	}

	return
}
