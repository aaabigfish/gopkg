// etcdV3 rebuild on go kit sd etcd
// add some more operator with etcd
package etcdv3

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var (
	// ErrNoKey indicates a client method needs a key but receives none.
	ErrNoKey = errors.New("no key provided")

	// ErrNoValue indicates a client method needs a value but receives none.
	ErrNoValue = errors.New("no value provided")

	// ErrNotfound
	ErrNotfound = errors.New("not found")
)

// KvEntry
type KvEntry struct {
	Key   string
	Value string
}

type WatchEvent struct {
	OpType int32
	Kv     *KvEntry
}

// Client is a wrapper around the etcd client.
type Client interface {
	GetEtcdClient() *clientv3.Client

	GetEtcdKV() clientv3.KV
	// GetEntries queries the given prefix in etcd and returns a slice
	// containing the values of all keys found, recursively, underneath that
	// prefix.
	GetEntries(prefix string) ([]*KvEntry, error)

	// Put save data
	Put(key, val string) (int64, int64, error)

	// PutJSON save interface
	PutJSON(key string, obj any) (int64, int64, error)

	// Get query data
	Get(key string) ([]string, error)

	Delete(key string, opts ...clientv3.OpOption) (int64, error)

	BatchDelete(keys []string, opts ...clientv3.OpOption) error

	// WatchPrefix watches the given prefix in etcd for changes. When a change
	// is detected, it will signal on the passed channel. Clients are expected
	// to call GetEntries to update themselves with the latest set of complete
	// values. WatchPrefix will always send an initial sentinel value on the
	// channel after establishing the watch, to ensure that clients always
	// receive the latest set of values. WatchPrefix will block until the
	// context passed to the NewClient constructor is terminated.
	WatchPrefix(prefix string, ev chan []*WatchEvent)

	// Register a service with etcd.
	Register(s Service, servCheckChan chan struct{}) error

	// Deregister a service with etcd.
	Deregister(s Service) error

	// LeaseID returns the lease id created for this service instance
	LeaseID() int64
}

const minHeartBeatTime = 500 * time.Millisecond

type EtcdEntity interface {
	Key() string
	Value() string
}

// Service holds the instance identifying data you want to publish to etcd. Key
// must be unique, and value is the string returned to subscribers, typically
// called the "instance" string in other parts of package sd.
type Service struct {
	Key   string // unique key, e.g. "/service/foobar/1.2.3.4:8080"
	Value string // returned to subscribers, e.g. "http://1.2.3.4:8080"
	TTL   *TTLOption
}

// TTLOption allow setting a key with a TTL. This option will be used by a loop
// goroutine which regularly refreshes the lease of the key.
type TTLOption struct {
	heartbeat time.Duration // e.g. time.Second * 3
	ttl       time.Duration // e.g. time.Second * 10
}

// NewTTLOption returns a TTLOption that contains proper TTL settings. Heartbeat
// is used to refresh the lease of the key periodically; its value should be at
// least 500ms. TTL defines the lease of the key; its value should be
// significantly greater than heartbeat.
//
// Good default values might be 3s heartbeat, 10s TTL.
func NewTTLOption(heartbeat, ttl time.Duration) *TTLOption {
	if heartbeat <= minHeartBeatTime {
		heartbeat = minHeartBeatTime
	}
	if ttl <= heartbeat {
		ttl = 3 * heartbeat
	}
	return &TTLOption{
		heartbeat: heartbeat,
		ttl:       ttl,
	}
}

type client struct {
	cli *clientv3.Client
	ctx context.Context

	kv clientv3.KV

	// Watcher interface instance, used to leverage Watcher.Close()
	watcher clientv3.Watcher
	// watcher context
	wctx context.Context
	// watcher cancel func
	wcf context.CancelFunc

	// leaseID will be 0 (clientv3.NoLease) if a lease was not created
	leaseID clientv3.LeaseID

	hbch <-chan *clientv3.LeaseKeepAliveResponse
	// Lease interface instance, used to leverage Lease.Close()
	leaser clientv3.Lease
}

// ClientOptions defines options for the etcd client. All values are optional.
// If any duration is not specified, a default of 3 seconds will be used.
type ClientOptions struct {
	Cert          string
	Key           string
	CACert        string
	DialTimeout   time.Duration
	DialKeepAlive time.Duration

	// DialOptions is a list of dial options for the gRPC client (e.g., for interceptors).
	// For example, pass grpc.WithBlock() to block until the underlying connection is up.
	// Without this, Dial returns immediately and connecting the server happens in background.
	DialOptions []grpc.DialOption

	Username string
	Password string
}

// NewClient returns Client with a connection to the named machines. It will
// return an error if a connection to the cluster cannot be made.
func NewClient(ctx context.Context, machines []string, options ClientOptions) (Client, error) {
	if options.DialTimeout == 0 {
		options.DialTimeout = 3 * time.Second
	}
	if options.DialKeepAlive == 0 {
		options.DialKeepAlive = 3 * time.Second
	}

	var err error
	var tlscfg *tls.Config

	if options.Cert != "" && options.Key != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      options.Cert,
			KeyFile:       options.Key,
			TrustedCAFile: options.CACert,
		}
		tlscfg, err = tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	cli, err := clientv3.New(clientv3.Config{
		Context:           ctx,
		Endpoints:         machines,
		DialTimeout:       options.DialTimeout,
		DialKeepAliveTime: options.DialKeepAlive,
		DialOptions:       options.DialOptions,
		TLS:               tlscfg,
		Username:          options.Username,
		Password:          options.Password,
	})
	if err != nil {
		return nil, err
	}

	return &client{
		cli: cli,
		ctx: ctx,
		kv:  clientv3.NewKV(cli),
	}, nil
}

func (c *client) GetEtcdClient() *clientv3.Client { return c.cli }

func (c *client) GetEtcdKV() clientv3.KV { return c.kv }

// LeaseID implements the etcd Client interface.
func (c *client) LeaseID() int64 { return int64(c.leaseID) }

// Put implements the etcd Client interface.
func (c *client) Put(key, val string) (int64, int64, error) {
	resp, err := c.kv.Put(c.ctx, key, val, clientv3.WithPrevKV())
	if err != nil {
		return 0, 0, err
	}

	if resp.PrevKv == nil {
		return 0, 0, nil
	}
	return resp.PrevKv.Version, resp.PrevKv.ModRevision, nil
}

func (c *client) PutJSON(key string, obj any) (int64, int64, error) {
	val, err := json.Marshal(obj)
	if err != nil {
		return 0, 0, err
	}
	return c.Put(key, string(val))
}

func (c *client) Delete(key string, opts ...clientv3.OpOption) (int64, error) {
	if key == "" {
		return 0, ErrNoKey
	}

	resp, err := c.cli.Delete(c.ctx, key, opts...)
	if err != nil {
		return 0, err
	}
	return resp.Deleted, nil
}

func (c *client) BatchDelete(keys []string, opts ...clientv3.OpOption) error {
	for i := range keys {
		if _, err := c.Delete(keys[i], opts...); err != nil {
			return fmt.Errorf("delete etcd key[%s] failed: %s", keys[i], err)
		}
	}
	return nil
}

// Get implements the etcd Client interface.
func (c *client) Get(key string) ([]string, error) {
	resp, err := c.kv.Get(c.ctx, key)
	if err != nil {
		return nil, err
	}
	sli := []string{}
	for _, o := range resp.Kvs {
		sli = append(sli, string(o.Value))
	}
	return sli, nil
}

// GetEntries implements the etcd Client interface.
func (c *client) GetEntries(key string) ([]*KvEntry, error) {
	resp, err := c.kv.Get(c.ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	entries := make([]*KvEntry, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		entry := &KvEntry{
			Key:   string(kv.Key),
			Value: string(kv.Value),
		}
		entries[i] = entry
	}
	return entries, nil
}

// WatchPrefix implements the etcd Client interface.
func (c *client) WatchPrefix(prefix string, ch chan []*WatchEvent) {
	c.wctx, c.wcf = context.WithCancel(c.ctx)
	c.watcher = clientv3.NewWatcher(c.cli)

	wch := c.watcher.Watch(c.wctx, prefix, clientv3.WithPrefix(), clientv3.WithRev(0))

	for wr := range wch {
		if wr.Canceled {
			return
		}
		evSlice := []*WatchEvent{}
		for _, ev := range wr.Events {
			wev := &WatchEvent{
				OpType: int32(ev.Type),
				Kv: &KvEntry{
					Key:   string(ev.Kv.Key),
					Value: string(ev.Kv.Value),
				},
			}
			evSlice = append(evSlice, wev)
		}
		ch <- evSlice
	}
}

func (c *client) Register(s Service, servCheckChan chan struct{}) error {
	if s.Key == "" {
		return ErrNoKey
	}
	if s.Value == "" {
		return ErrNoValue
	}

	if c.leaser != nil {
		c.leaser.Close()
	}
	c.leaser = clientv3.NewLease(c.cli)

	if c.watcher != nil {
		c.watcher.Close()
	}
	c.watcher = clientv3.NewWatcher(c.cli)
	if c.kv == nil {
		c.kv = clientv3.NewKV(c.cli)
	}

	if s.TTL == nil {
		s.TTL = NewTTLOption(time.Second*3, time.Second*10)
	}

	grantResp, err := c.leaser.Grant(c.ctx, int64(s.TTL.ttl.Seconds()))
	if err != nil {
		return err
	}
	c.leaseID = grantResp.ID

	_, err = c.kv.Put(
		c.ctx,
		s.Key,
		s.Value,
		clientv3.WithLease(c.leaseID),
	)
	if err != nil {
		return err
	}

	// this will keep the key alive 'forever' or until we revoke it or
	// the context is canceled
	c.hbch, err = c.leaser.KeepAlive(c.ctx, c.leaseID)
	if err != nil {
		return err
	}

	// discard the keepalive response, make etcd library not to complain
	// fix bug #799
	go func() {
		for {
			select {
			case r := <-c.hbch:
				// avoid dead loop when channel was closed
				if r == nil {
					servCheckChan <- struct{}{}
					return
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (c *client) Deregister(s Service) error {
	defer c.close()

	if s.Key == "" {
		return ErrNoKey
	}
	if _, err := c.cli.Delete(c.ctx, s.Key, clientv3.WithIgnoreLease()); err != nil {
		return err
	}

	return nil
}

// close will close any open clients and call
// the watcher cancel func
func (c *client) close() {
	if c.leaser != nil {
		c.leaser.Close()
	}
	if c.watcher != nil {
		c.watcher.Close()
	}
	if c.wcf != nil {
		c.wcf()
	}
}
