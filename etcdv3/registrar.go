package etcdv3

import (
	"sync"
	"time"

	"gitlab.ipcloud.cc/go/gopkg/log"
)

// Registrar registers service instance liveness information to etcd.
type Registrar struct {
	client  Client
	service Service

	quitmtx   sync.Mutex
	checkServ chan struct{}
	quit      chan struct{}
}

// NewRegistrar returns a etcd Registrar acting on the provided catalog
// registration (service).
func NewRegistrar(client Client, service Service) *Registrar {
	return &Registrar{
		client:    client,
		service:   service,
		checkServ: make(chan struct{}, 1),
	}
}

// Register implements the sd.Registrar interface. Call it when you want your
// service to be registered in etcd, typically at startup.
func (r *Registrar) Register() error {
	if err := r.client.Register(r.service, r.checkServ); err != nil {
		log.Errorf("Register(%+v) err(%v)", r, err)
		return err
	}
	if r.service.TTL != nil {
		log.Warnf("Register action lease(%+v)", r.client.LeaseID())
	}

	go r.AutoCheckService()
	return nil
}

func (r *Registrar) AutoCheckService() {
	for {
		select {
		case <-r.checkServ:
			log.Warn("AutoCheckService repeat register")
			err := r.Register()
			if err != nil {
				time.AfterFunc(3*time.Second, func() {
					r.checkServ <- struct{}{}
				})
			}
		}
	}
}

// Deregister implements the sd.Registrar interface. Call it when you want your
// service to be deregistered from etcd, typically just prior to shutdown.
func (r *Registrar) Deregister() {
	if err := r.client.Deregister(r.service); err != nil {
		log.Errorf("Deregister(%+v) err(%v)", r.service, err)
	} else {
		log.Info("action deregister")
	}

	r.quitmtx.Lock()
	defer r.quitmtx.Unlock()
	if r.quit != nil {
		close(r.quit)
		r.quit = nil
	}
}
