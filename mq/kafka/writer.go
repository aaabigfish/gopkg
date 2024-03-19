package kafka

import (
	"context"
	"time"

	"github.com/rs/xid"
	kf "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/snappy"
)

type Writer interface {
	WriteMessage(tx context.Context, msg Message) error
	WriteMessages(tx context.Context, msgs []Message) error
	GetWriter() *kf.Writer
	Close() error
}
type writer struct {
	kfWriter *kf.Writer
	c        *kf.WriterConfig
}

func NewWriter(brokers []string, topic ...string) Writer {
	dialer := &kf.Dialer{
		Timeout:  10 * time.Second,
		ClientID: xid.New().String(),
	}

	conf := kf.WriterConfig{
		Brokers:          brokers,
		Async:            true,
		Balancer:         &kf.Hash{},
		Dialer:           dialer,
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      10 * time.Second,
		CompressionCodec: snappy.NewCompressionCodec(),
	}

	if len(topic) > 0 {
		conf.Topic = topic[0]
	}

	return &writer{kfWriter: kf.NewWriter(conf), c: &conf}
}

func (w *writer) GetWriter() *kf.Writer {
	return w.kfWriter
}

func (w *writer) WriteMessage(tx context.Context, msg Message) error {
	return w.kfWriter.WriteMessages(tx, msg)
}

func (w *writer) WriteMessages(tx context.Context, msgs []Message) error {
	return w.kfWriter.WriteMessages(tx, msgs...)
}

func (w *writer) Close() error {
	return w.kfWriter.Close()
}
