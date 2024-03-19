package etcdv3

import "sync"

type Cache struct {
	mtx       sync.RWMutex
	Instances []*KvEntry
}

func (c *Cache) UpdateByKvPair(kv []*KvEntry) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.Instances = nil
	c.Instances = make([]*KvEntry, len(kv))
	for i, v := range kv {
		c.Instances[i] = &KvEntry{
			Key:   v.Key,
			Value: v.Value,
		}
	}
}

func (c *Cache) UpdateByWatchEvent(wev []*WatchEvent) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	for _, e := range wev {
		if e.OpType == 0 {
			kv, _ := c.GetKvEntry(e)
			if kv == nil {
				c.Instances = append(c.Instances, e.Kv)
			} else {
				kv.Value = e.Kv.Value
			}
		} else {
			_, i := c.GetKvEntry(e)
			if i == 0 && len(c.Instances) == 1 {
				c.Instances = nil
			} else {
				c.Instances = append(c.Instances[:i], c.Instances[i+1:]...)
			}
		}
	}
}

func (c *Cache) GetKvEntry(ev *WatchEvent) (*KvEntry, int) {
	for i, o := range c.Instances {
		if o.Key == ev.Kv.Key {
			return c.Instances[i], i
		}
	}
	return nil, -1
}
