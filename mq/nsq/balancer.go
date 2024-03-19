package nsq

import (
	"hash"
	"hash/crc32"
	"hash/fnv"
	"math/rand"
	"sync"
	"sync/atomic"
)

// The Balancer interface provides an abstraction of the message distribution
// logic used by Writer instances to route messages to the partitions available
// on a nsq cluster.
//
// Balancers must be safe to use concurrently from multiple goroutines.
type Balancer interface {
	// Balance receives a message and a set of available partitions and
	// returns the partition number that the message should be routed to.
	//
	// An application should refrain from using a balancer to manage multiple
	// sets of partitions (from different topics for examples), use one balancer
	// instance for each partition set, so the balancer can detect when the
	// partitions change and assume that the nsq topic has been rebalanced.
	Balance(key []byte, partitions ...int) (partition int)
}

// BalancerFunc is an implementation of the Balancer interface that makes it
// possible to use regular functions to distribute messages across partitions.
type BalancerFunc func([]byte, ...int) int

// Balance calls f, satisfies the Balancer interface.
func (f BalancerFunc) Balance(key []byte, partitions ...int) int {
	return f(key, partitions...)
}

// RoundRobin is an Balancer implementation that equally distributes messages
// across all available partitions.
type RoundRobin struct {
	// Use a 32 bits integer so RoundRobin values don't need to be aligned to
	// apply atomic increments.
	offset uint32
}

// Balance satisfies the Balancer interface.
func (rr *RoundRobin) Balance(key []byte, partitions ...int) int {
	return rr.balance(partitions)
}

func (rr *RoundRobin) balance(partitions []int) int {
	length := uint32(len(partitions))
	offset := atomic.AddUint32(&rr.offset, 1) - 1
	return partitions[offset%length]
}

var (
	fnv1aPool = &sync.Pool{
		New: func() interface{} {
			return fnv.New32a()
		},
	}
)

// Hash is a Balancer that uses the provided hash function to determine which
// partition to route messages to.  This ensures that messages with the same key
// are routed to the same partition.
//
// The logic to calculate the partition is:
//
//	hasher.Sum32() % len(partitions) => partition
//
// By default, Hash uses the FNV-1a algorithm.  This is the same algorithm used
// by the Sarama Producer and ensures that messages produced by nsq-go will
// be delivered to the same topics that the Sarama producer would be delivered to.
type Hash struct {
	rr     RoundRobin
	Hasher hash.Hash32

	// lock protects Hasher while calculating the hash code.  It is assumed that
	// the Hasher field is read-only once the Balancer is created, so as a
	// performance optimization, reads of the field are not protected.
	lock sync.Mutex
}

func (h *Hash) Balance(key []byte, partitions ...int) int {
	if key == nil {
		return h.rr.Balance(key, partitions...)
	}

	hasher := h.Hasher
	if hasher != nil {
		h.lock.Lock()
		defer h.lock.Unlock()
	} else {
		hasher = fnv1aPool.Get().(hash.Hash32)
		defer fnv1aPool.Put(hasher)
	}

	hasher.Reset()
	if _, err := hasher.Write(key); err != nil {
		panic(err)
	}

	// uses same algorithm that Sarama's hashPartitioner uses
	// note the type conversions here.  if the uint32 hash code is not cast to
	// an int32, we do not get the same result as sarama.
	partition := int32(hasher.Sum32()) % int32(len(partitions))
	if partition < 0 {
		partition = -partition
	}

	return int(partition)
}

// ReferenceHash is a Balancer that uses the provided hash function to determine which
// partition to route messages to.  This ensures that messages with the same key
// are routed to the same partition.
//
// The logic to calculate the partition is:
//
//	(int32(hasher.Sum32()) & 0x7fffffff) % len(partitions) => partition
//
// By default, ReferenceHash uses the FNV-1a algorithm. This is the same algorithm as
// the Sarama NewReferenceHashPartitioner and ensures that messages produced by nsq-go will
// be delivered to the same topics that the Sarama producer would be delivered to.
type ReferenceHash struct {
	rr     randomBalancer
	Hasher hash.Hash32

	// lock protects Hasher while calculating the hash code.  It is assumed that
	// the Hasher field is read-only once the Balancer is created, so as a
	// performance optimization, reads of the field are not protected.
	lock sync.Mutex
}

func (h *ReferenceHash) Balance(key []byte, partitions ...int) int {
	if key == nil {
		return h.rr.Balance(key, partitions...)
	}

	hasher := h.Hasher
	if hasher != nil {
		h.lock.Lock()
		defer h.lock.Unlock()
	} else {
		hasher = fnv1aPool.Get().(hash.Hash32)
		defer fnv1aPool.Put(hasher)
	}

	hasher.Reset()
	if _, err := hasher.Write(key); err != nil {
		panic(err)
	}

	// uses the same algorithm as the Sarama's referenceHashPartitioner.
	// note the type conversions here. if the uint32 hash code is not cast to
	// an int32, we do not get the same result as sarama.
	partition := (int32(hasher.Sum32()) & 0x7fffffff) % int32(len(partitions))
	return int(partition)
}

type randomBalancer struct {
	mock int // mocked return value, used for testing
}

func (b randomBalancer) Balance(key []byte, partitions ...int) (partition int) {
	if b.mock != 0 {
		return b.mock
	}
	return partitions[rand.Int()%len(partitions)]
}

// CRC32Balancer is a Balancer that uses the CRC32 hash function to determine
// which partition to route messages to.  This ensures that messages with the
// same key are routed to the same partition.  This balancer is compatible with
// the built-in hash partitioners in librdnsq and the language bindings that
// are built on top of it, including the
// github.com/confluentinc/confluent-kafka-go Go package.
//
// With the Consistent field false (default), this partitioner is equivalent to
// the "consistent_random" setting in librdkafka.  When Consistent is true, this
// partitioner is equivalent to the "consistent" setting.  The latter will hash
// empty or nil keys into the same partition.
//
// Unless you are absolutely certain that all your messages will have keys, it's
// best to leave the Consistent flag off.  Otherwise, you run the risk of
// creating a very hot partition.
type CRC32Balancer struct {
	Consistent bool
	random     randomBalancer
}

func (b CRC32Balancer) Balance(key []byte, partitions ...int) (partition int) {
	// NOTE: the crc32 balancers in librdkafka don't differentiate between nil
	//       and empty keys.  both cases are treated as unset.
	if len(key) == 0 && !b.Consistent {
		return b.random.Balance(key, partitions...)
	}

	idx := crc32.ChecksumIEEE(key) % uint32(len(partitions))
	return partitions[idx]
}

// Murmur2Balancer is a Balancer that uses the Murmur2 hash function to
// determine which partition to route messages to.  This ensures that messages
// with the same key are routed to the same partition.  This balancer is
// compatible with the partitioner used by the Java library and by librdkafka's
// "murmur2" and "murmur2_random" partitioners.
//
// With the Consistent field false (default), this partitioner is equivalent to
// the "murmur2_random" setting in librdkafka.  When Consistent is true, this
// partitioner is equivalent to the "murmur2" setting.  The latter will hash
// nil keys into the same partition.  Empty, non-nil keys are always hashed to
// the same partition regardless of configuration.
//
// Unless you are absolutely certain that all your messages will have keys, it's
// best to leave the Consistent flag off.  Otherwise, you run the risk of
// creating a very hot partition.
//
// Note that the librdkafka documentation states that the "murmur2_random" is
// functionally equivalent to the default Java partitioner.  That's because the
// Java partitioner will use a round robin balancer instead of random on nil
// keys.  We choose librdkafka's implementation because it arguably has a larger
// install base.
type Murmur2Balancer struct {
	Consistent bool
	random     randomBalancer
}

func (b Murmur2Balancer) Balance(key []byte, partitions ...int) (partition int) {
	// NOTE: the murmur2 balancers in java and librdkafka treat a nil key as
	//       non-existent while treating an empty slice as a defined value.
	if key == nil && !b.Consistent {
		return b.random.Balance(key, partitions...)
	}

	idx := (murmur2(key) & 0x7fffffff) % uint32(len(partitions))
	return partitions[idx]
}

// Go port of the Java library's murmur2 function.
// https://github.com/apache/kafka/blob/1.0/clients/src/main/java/org/apache/kafka/common/utils/Utils.java#L353
func murmur2(data []byte) uint32 {
	length := len(data)
	const (
		seed uint32 = 0x9747b28c
		// 'm' and 'r' are mixing constants generated offline.
		// They're not really 'magic', they just happen to work well.
		m = 0x5bd1e995
		r = 24
	)

	// Initialize the hash to a random value
	h := seed ^ uint32(length)
	length4 := length / 4

	for i := 0; i < length4; i++ {
		i4 := i * 4
		k := (uint32(data[i4+0]) & 0xff) + ((uint32(data[i4+1]) & 0xff) << 8) + ((uint32(data[i4+2]) & 0xff) << 16) + ((uint32(data[i4+3]) & 0xff) << 24)
		k *= m
		k ^= k >> r
		k *= m
		h *= m
		h ^= k
	}

	// Handle the last few bytes of the input array
	extra := length % 4
	if extra >= 3 {
		h ^= (uint32(data[(length & ^3)+2]) & 0xff) << 16
	}
	if extra >= 2 {
		h ^= (uint32(data[(length & ^3)+1]) & 0xff) << 8
	}
	if extra >= 1 {
		h ^= uint32(data[length & ^3]) & 0xff
		h *= m
	}

	h ^= h >> 13
	h *= m
	h ^= h >> 15

	return h
}

var nodesCache atomic.Value

func loadCachedNodes(numNodes int) []int {
	nodes, ok := nodesCache.Load().([]int)
	if ok && len(nodes) >= numNodes {
		return nodes[:numNodes]
	}

	const alignment = 128
	n := ((numNodes / alignment) + 1) * alignment

	nodes = make([]int, n)
	for i := range nodes {
		nodes[i] = i
	}

	nodesCache.Store(nodes)
	return nodes[:numNodes]
}
