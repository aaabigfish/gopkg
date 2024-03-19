package kafka

import (
	"context"
	"time"

	kf "github.com/segmentio/kafka-go"
)

// 注意消费者的kafka版本必须0.10以上的版本
type Reader interface {
	ReadMessage(ctx context.Context) (Message, error)
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msgs ...Message) error
	GetReader() *kf.Reader
	AddHook(...ReaderFunc)
	Do(m Message) error
	Close() error
}

type reader struct {
	kfReader *kf.Reader
	hooks    []ReaderFunc
}

func NewReader(brokers []string, topic string, groupId string) Reader {
	dialer := &kf.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// make a new reader that consumes from topic-A
	r := kf.NewReader(kf.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupId,
		Dialer:   dialer,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		// CommitInterval: 10 * time.Second, // flushes commits to Kafka every second
	})

	return &reader{
		kfReader: r,
		hooks:    make([]ReaderFunc, 0),
	}
}

func (r *reader) GetReader() *kf.Reader {
	return r.kfReader
}

func (r *reader) ReadMessage(ctx context.Context) (Message, error) {
	return r.kfReader.ReadMessage(ctx)
}

func (r *reader) FetchMessage(ctx context.Context) (Message, error) {
	return r.kfReader.FetchMessage(ctx)
}

func (r *reader) CommitMessages(ctx context.Context, msgs ...Message) error {
	return r.kfReader.CommitMessages(ctx, msgs...)
}

func (r *reader) AddHook(hook ...ReaderFunc) {
	r.hooks = append(r.hooks, hook...)
}

func (r *reader) Do(m Message) error {
	for _, hook := range r.hooks {
		if err := hook(m.Partition, m.Offset, m.Key, m.Value); err != nil {
			return err
		}
	}

	return nil
}

func (r *reader) Close() error {
	return r.kfReader.Close()
}
