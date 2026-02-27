package lrukreplacer

import (
	"errors"
)

// represents metadata of a key in LRU-K cache
type node struct {
	key      int     // the value of key in cache
	value    int     // the corresponding value to be stored for a key
	register []int64 // timestamp register - to be maintained when key is accessed (it should act as a queue)
}

func NewNode(key, val, k int) *node {
	return &node{
		key:      key,
		value:    val,
		register: make([]int64, 0, k),
	}
}

func (n *node) validate() error {
	if n == nil || n.register == nil {
		return errors.New("invalid node")
	}

	return nil
}

func (n *node) Get() (int, []int64, error) {
	if err := n.validate(); err != nil {
		return 0, nil, err
	}

	return n.value, n.register, nil
}

func (n *node) RecordAccess(timestamp int64) error {
	if err := n.validate(); err != nil {
		return err
	}

	if len(n.register) == cap(n.register) {
		n.register = n.register[1:]
	}

	n.register = append(n.register, timestamp)
	return nil
}
