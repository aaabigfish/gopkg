package asynq

import (
	"sync"

	"github.com/hibiken/asynq"
)

type Reader interface {
	GetReader() *asynq.Server
	AddHook(string, HandlerFunc)
	AddHooks(Handler)
	Run() error
	Close()
}

type reader struct {
	c      *RedisClientOpt
	reader *asynq.Server
	mux    *asynq.ServeMux
	hooks  Handler
	rwl    sync.RWMutex
}

func NewReader(opt *RedisClientOpt, c ...asynq.Config) Reader {
	cfg := asynq.Config{}

	if len(c) > 0 {
		cfg = c[0]
	}
	r := asynq.NewServer(opt, cfg)

	return &reader{
		c:      opt,
		reader: r,
		mux:    asynq.NewServeMux(),
		hooks:  make(map[string]HandlerFunc, 10),
	}
}

func (r *reader) GetReader() *asynq.Server {
	return r.reader
}

func (r *reader) AddHook(pattern string, hookFunc HandlerFunc) {
	r.rwl.Lock()
	r.hooks[pattern] = hookFunc
	r.rwl.Unlock()
}

func (r *reader) AddHooks(hooks Handler) {
	for pattern, hookFunc := range hooks {
		if hookFunc == nil {
			continue
		}

		r.rwl.Lock()
		r.hooks[pattern] = hookFunc
		r.rwl.Unlock()
	}
}

func (r *reader) Run() error {
	for pattern, handler := range r.hooks {
		r.mux.HandleFunc(pattern, handler)
	}

	if err := r.reader.Run(r.mux); err != nil {
		return err
	}

	return nil
}

func (r *reader) Close() {
	if r.reader == nil {
		r.reader.Shutdown()
		return
	}
}
