package cacheNode

import (
	"errors"
)

// represents metadata of a key in LRU-K cache
type Node struct {
	Key      int     // the value of key in cache
	Value    int     // the corresponding value to be stored for a key
	Register []int64 // timestamp register - to be maintained when key is accessed (it should act as a queue)
}

func NewNode(key, val, k int) *Node {
	return &Node{
		Key:      key,
		Value:    val,
		Register: make([]int64, 0, k),
	}
}

func (n *Node) Validate() error {
	if n == nil || n.Register == nil {
		return errors.New("invalid node")
	}

	return nil
}

// func (n *node) Get() (int, []int64, error) {
// 	if err := n.validate(); err != nil {
// 		return 0, nil, err
// 	}

// 	return n.value, n.register, nil
// }

func (n *Node) RecordAccess(timestamp int64) error {
	if err := n.Validate(); err != nil {
		return err
	}

	if len(n.Register) == cap(n.Register) {
		n.Register = n.Register[1:]
	}

	n.Register = append(n.Register, timestamp)
	return nil
}
