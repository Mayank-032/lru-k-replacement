package lrukreplacer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
)

type cache struct {
	capacity                   int
	timestampsRegisterCapacity int // denotes the k timestamps need to be maintained for a key
	data                       map[int]*node
	timestamp                  int64

	mu sync.Mutex
}

var _ ICache = (*cache)(nil)

func InitCache(cacheCapacity, timestampsRegisterCapacity int) *cache {
	var c = &cache{
		capacity:                   cacheCapacity,
		timestampsRegisterCapacity: timestampsRegisterCapacity,
		data:                       make(map[int]*node, cacheCapacity),
	}

	return c
}

func (c *cache) Get(key int) (int, error) {
	fmt.Println()

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.validate(); err != nil {
		return -1, err
	}

	c.timestamp = c.timestamp + 1

	var node *node
	var ok bool

	node, ok = c.data[key]

	if !ok {
		return -1, errors.New("key does not exists in cache")
	}

	if err := node.RecordAccess(c.timestamp); err != nil {
		return -1, err
	}

	dataB, _ := json.Marshal(c.data)
	fmt.Println("cache: ", string(dataB))

	return node.value, nil
}

func (c *cache) Set(key, val int) (int, error) {
	fmt.Println()

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.validate(); err != nil {
		return -1, err
	}

	c.timestamp = c.timestamp + 1

	var node *node
	var ok bool

	node, ok = c.data[key]
	if ok {
		node.value = val
	} else {
		if c.isFull() {
			err := c.evict()
			if err != nil {
				return -1, err
			}
		}

		node = NewNode(key, val, c.timestampsRegisterCapacity)
	}

	if err := node.RecordAccess(c.timestamp); err != nil {
		return -1, err
	}

	c.data[node.key] = node

	dataB, _ := json.Marshal(c.data)
	fmt.Println("cache: ", string(dataB))

	return key, nil
}

func (c *cache) validate() error {
	if c == nil || c.data == nil {
		return errors.New("cache is not initialized")
	}

	return nil
}

func (c *cache) isFull() bool {
	if len(c.data) == c.capacity {
		return true
	}
	return false
}

func (c *cache) evict() error {
	// first check for elements in cache which are not accessed atleast k times

	var (
		minMTKTimesAccessed     int64 = int64(math.MaxInt64)
		minNodeMTKTimesAccessed *node

		minLTKTimesAccessed     int64 = int64(math.MaxInt64)
		minNodeLTKTimesAccessed *node
	)

	for _, node := range c.data {
		if len(node.register) != cap(node.register) {
			if node.register[0] < minLTKTimesAccessed {
				minLTKTimesAccessed = node.register[0]
				minNodeLTKTimesAccessed = node
			}
		} else {
			if node.register[0] < minMTKTimesAccessed {
				minMTKTimesAccessed = node.register[0]
				minNodeMTKTimesAccessed = node
			}
		}
	}

	if minLTKTimesAccessed != 0 && minNodeLTKTimesAccessed != nil {
		delete(c.data, minNodeLTKTimesAccessed.key)
		return nil
	}

	if minMTKTimesAccessed != 0 && minNodeMTKTimesAccessed != nil {
		delete(c.data, minNodeMTKTimesAccessed.key)
		return nil
	}

	return errors.New("unable to evict from cache")
}
