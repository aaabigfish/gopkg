package ecode

import (
	"fmt"
)

type Message struct {
	ECode
	EMsg string
}

// Error new message with code and msg
func Error(code ECode, msg string) *Message {
	return &Message{ECode: code, EMsg: msg}
}

// Errorf new message with code and msg
func Errorf(code ECode, format string, args ...interface{}) *Message {
	return Error(code, fmt.Sprintf(format, args...))
}

func (m *Message) String() string {
	if m.EMsg == "" {
		return m.ECode.String()
	}
	return fmt.Sprintf("%s:%s", m.ECode.String(), m.EMsg)
}

func (m *Message) Error() string {
	return m.String()
}
