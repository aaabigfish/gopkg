package kafka

import (
	kf "github.com/segmentio/kafka-go"
)

type Message = kf.Message
type ReaderFunc func(partition int, offset int64, key []byte, val []byte) error
type WriterFunc func(string, string) Writer
