package etcdv3

import (
	"github.com/aaabigfish/gopkg/log"
)

// Instancer yields instances stored in a certain etcd keyspace. Any kind of
// change in that keyspace is watched and will update the Instancer's Instancers.
type Instancer struct {
	cache    *Cache
	client   Client
	prefix   string
	quitc    chan struct{}
	callback func([]*WatchEvent) error
}

// NewInstancer returns an etcd instancer. It will start watching the given
// prefix for changes, and update the subscribers.
func NewInstancer(c Client, prefix string) (*Instancer, error) {
	s := &Instancer{
		client: c,
		prefix: prefix,
		cache:  &Cache{},
		quitc:  make(chan struct{}),
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err == nil {
		log.Info("NewInstancer prefix(%+v) instancesLen(%v)", s.prefix, len(instances))
	} else {
		log.Errorf("NewInstancer prefix(%+v) err(%v)", s.prefix, err)
	}
	s.cache.UpdateByKvPair(instances)

	// routine loop for watch
	go s.loop()
	return s, nil
}

func (s *Instancer) loop() {
	ch := make(chan []*WatchEvent)
	go s.client.WatchPrefix(s.prefix, ch)

	for {
		select {
		case wev := <-ch:
			s.cache.UpdateByWatchEvent(wev)
			if s.callback != nil {
				s.callback(wev)
			}
		case <-s.quitc:
			return
		}
	}
}

// Subscribe if watch event will notice caller.
func (s *Instancer) Subscribe(callback func([]*WatchEvent) error) {
	s.callback = callback
}

func (s *Instancer) Cached() *Cache {
	return s.cache
}

// Stop terminates the Instancer.
func (s *Instancer) Stop() {
	close(s.quitc)
}
