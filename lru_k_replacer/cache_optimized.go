package lrukreplacer

import (
	"container/heap"
	"errors"
	"log"
	"lruKReplacer/pkg/utils"
	"sync"
)

type optimizedCache struct {
	capacity                   int
	timestampsRegisterCapacity int   // denotes the k timestamps need to be maintained for a key
	timestamp                  int64 // global timestamp counter

	// map of key-value pairs: denotes the registry of nodes and its address in the cache
	data map[int]*node

	// both nodes head and tail keeps track of nodes whose occurrence is less than timestampsRegisterCapacity
	head *utils.DllNode // head pointer - points to the node is which had most recent access
	tail *utils.DllNode // tail pointer - points to the node is which has oldest access

	// keeps track of nodes whose occurrence is greater than or equal to timestampsRegisterCapacity
	minHeap *PriorityQueue

	mu sync.Mutex
}

var _ ICache = (*optimizedCache)(nil)

func InitOptimizedCache(cacheCapacity, timestampsRegisterCapacity int) *optimizedCache {
	pq := &PriorityQueue{}
	heap.Init(pq)

	var c = &optimizedCache{
		capacity:                   cacheCapacity,
		timestampsRegisterCapacity: timestampsRegisterCapacity,
		data:                       make(map[int]*node, cacheCapacity),
		minHeap:                    pq,
	}

	return c
}

func (oc *optimizedCache) Get(key int) (int, error) {
	// validate cache
	if err := oc.validate(); err != nil {
		return -1, err
	}

	// taking a lock
	oc.mu.Lock()
	defer oc.mu.Unlock()

	// increase global timestamp counter
	oc.timestamp = oc.timestamp + 1

	// checking if node exists or not
	node, ok := oc.data[key]
	if !ok {
		// if not exists and since its get operation, return nil
		return -1, errors.New("key does not exists in cache")
	}

	// record the new timestamp
	if err := node.RecordAccess(oc.timestamp); err != nil {
		return -1, err
	}

	// now since the timestamp is recorded, we need to check whether it should be kept in dll or
	// we should move it to heap
	if len(node.register) == cap(node.register) {
		// if the capacity is reached, and after reaching capacity we are accessing it first time, just
		// remove from dll otherwise no operation on dll to be performed
		if node.dllNode != nil {
			dllNode := node.dllNode
			dllNode.Remove()
			node.dllNode = nil

			heap.Push(oc.minHeap, &Item{Node: node})
		} else {
			// push the node to heap
			heap.Fix(oc.minHeap, node.heapIndex)
		}
	} else {
		// if the capacity is not reached simple move the node to top as its the hot node for now
		// in dll
		dllNode := node.dllNode
		dllNode.Remove()
		dllNode.Insert(oc.head)
	}

	return node.value, nil
}

func (oc *optimizedCache) Set(key int, val int) (int, error) {
	// validate the cache
	if err := oc.validate(); err != nil {
		return -1, err
	}

	// taking the lock for this operation
	oc.mu.Lock()
	defer oc.mu.Unlock()

	// increase global timestamp counter
	oc.timestamp = oc.timestamp + 1

	// checking if node exists or not
	node, ok := oc.data[key]
	if ok {
		// if an existing node, just update the value
		node.value = val

		// record the new access in the node
		if err := node.RecordAccess(oc.timestamp); err != nil {
			return -1, err
		}

		// if the capacity is reached, we should perform similar operation as GET (with
		// perspective to DLL and Heap)
		if len(node.register) == cap(node.register) {
			if node.dllNode != nil {
				dllNode := node.dllNode
				dllNode.Remove()
				node.dllNode = nil

				heap.Push(oc.minHeap, &Item{Node: node})
			} else {
				// push the node to minheap
				heap.Fix(oc.minHeap, node.heapIndex)
			}
		} else {
			dllNode := node.dllNode
			dllNode.Remove()
			dllNode.Insert(oc.head)
		}
	} else {
		// if cache is full, based on the dll and heap we should make a decision to evict from the cache
		if oc.isFull() {
			err := oc.evict()
			if err != nil {
				return -1, err
			}
		}

		// since the node is not an existing one, create a new node
		node = NewNode(key, val, oc.timestampsRegisterCapacity)

		// record the new access in the node
		if err := node.RecordAccess(oc.timestamp); err != nil {
			return -1, err
		}

		// create a new dllNode and capture it in cache - because every new node will land in dll first
		dllNode := utils.NewDLLNode(node)
		dllNode.Insert(oc.head)

		node.dllNode = dllNode
	}

	// assign the key in the cache
	oc.data[key] = node

	return key, nil
}

func (oc *optimizedCache) validate() error {
	if oc == nil || oc.data == nil {
		return errors.New("cache is not initialized")
	}

	if oc.capacity == 0 || oc.timestampsRegisterCapacity == 0 {
		log.Printf("cache_capacity: %v, timestampsRegisterCapacity: %v\n", oc.capacity, oc.timestampsRegisterCapacity)
		return errors.New("capacities are not initialized in cache")
	}

	if oc.head == nil || oc.tail == nil {
		return errors.New("doublyLinkedList is not initialized in cache")
	}

	if oc.minHeap == nil {
		return errors.New("heap is not initialized in cache")
	}

	return nil
}

func (oc *optimizedCache) isFull() bool {
	if len(oc.data) == oc.capacity {
		return true
	}
	return false
}

func (oc *optimizedCache) evict() error {
	// first check if the nodes are still present in dll, those are not the proven nodes, hence should be
	// removed from cache as well as dll
	if oc.head.Next != oc.tail {
		dllNode := oc.tail.Prev
		dllNode.Remove()

		node, ok := dllNode.Value.(*node)
		if !ok || node == nil {
			return errors.New("invalid dll node")
		}

		delete(oc.data, node.key)
		return nil
	}

	// if no node is present in dll then remove the node from heap as well as cache
	element := heap.Pop(oc.minHeap)
	pqItem := element.(*Item)

	// also delete the node from cache which got polled
	delete(oc.data, pqItem.Node.key)

	return nil
}
