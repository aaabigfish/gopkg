package asynq

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/hibiken/asynq"
)

var (
	once   sync.Once
	client *asynq.Client
)

type Client = asynq.Client
type TaskInfo = asynq.TaskInfo

type Writer interface {
	Send(key string, val interface{}, opts ...asynq.Option) (*TaskInfo, error)
	Push(val interface{}, opts ...asynq.Option) (*TaskInfo, error)
	GetWriter() *asynq.Client
	Close() error
}

type writer struct {
	client *Client
	c      *RedisClientOpt
	topic  string
}

func NewWriter(con *RedisClientOpt, topic ...string) Writer {
	w := &writer{
		c: con,
	}
	once.Do(func() {
		client = asynq.NewClient(w.c)
	})
	w.client = client

	if len(topic) > 0 {
		w.topic = topic[0]
	}

	return w
}

func (w *writer) GetWriter() *Client {
	return w.client
}

func (w *writer) Send(key string, val interface{}, opts ...asynq.Option) (*TaskInfo, error) {
	if w.client == nil {
		return nil, errors.New("connection is closed")
	}

	payload, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(key, payload, opts...)

	return w.client.Enqueue(task)
}

func (w *writer) Push(val interface{}, opts ...asynq.Option) (*TaskInfo, error) {
	if w.topic == "" {
		return nil, errors.New("topic is empty")
	}

	return w.Send(w.topic, val, opts...)
}

func (w *writer) Close() error {
	if w.client != nil {
		return w.client.Close()
	}
	return nil
}
