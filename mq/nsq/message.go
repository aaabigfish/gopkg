package nsq

import (
	"encoding/json"
	"time"
)

type Header struct {
	Key   string
	Value []byte
}

type Message struct {
	Topic   string
	Headers []Header
	Key     []byte
	Body    []byte

	// If not set at the creation, Time will be automatically set when
	// writing the message.
	Time int64
}

func NewMessage() *Message {
	return &Message{
		Time: time.Now().UnixNano(),
	}
}

func (m *Message) GetKey() string {
	return string(m.Key)
}

func (m *Message) GetBody() []byte {
	return m.Body
}

func (m *Message) SetKey(key []byte) *Message {
	m.Key = key
	return m
}

func (m *Message) SetHeader(key string, val []byte) *Message {
	m.Headers = append(m.Headers, Header{
		Key:   key,
		Value: val,
	})
	return m
}

func (m *Message) SetTopic(topic string) *Message {
	m.Topic = topic
	return m
}

func (m *Message) SetBody(body []byte) *Message {
	m.Body = body
	return m
}

func (m *Message) SetBodyJSON(body any) *Message {
	data, _ := json.Marshal(body)

	return m.SetBody(data)
}
