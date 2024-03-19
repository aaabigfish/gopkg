package asynq

import (
	"context"

	"github.com/hibiken/asynq"
)

type Message = asynq.Task

type RedisClientOpt = asynq.RedisClientOpt

type HandlerFunc func(context.Context, *Message) error

type Handler map[string]HandlerFunc
