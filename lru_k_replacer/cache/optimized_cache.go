package cache

import (
	"container/heap"
	"errors"
	"log"
	cacheNode "lruKReplacer/lru_k_replacer/node"
	"lruKReplacer/pkg/utils"
	"sync"
)

type cacheEntry struct {
	node      *cacheNode.Node
	HeapIndex int
	DllNode   *utils.DllNode
}

type optimizedCache struct {
	capacity                   int
	timestampsRegisterCapacity int   // denotes the k timestamps need to be maintained for a key
	timestamp                  int64 // global timestamp counter

	// map of key-value pairs: denotes the registry of nodes and its address in the cache
	data map[int]*cacheEntry

	// both nodes head and tail keeps track of nodes whose occurrence is less than timestampsRegisterCapacity
	head *utils.DllNode // head pointer - points to the node is which had most recent access
	tail *utils.DllNode // tail pointer - points to the node is which has oldest access

	// keeps track of nodes whose occurrence is greater than or equal to timestampsRegisterCapacity
	minHeap *utils.PriorityQueue[cacheEntry]

	mu sync.Mutex
}

var _ ICache = (*optimizedCache)(nil)

func InitOptimizedCache(cacheCapacity, timestampsRegisterCapacity int) *optimizedCache {
	pq := &utils.PriorityQueue[cacheEntry]{
		LessFunc: func(a, b *cacheEntry) bool {
			return a.node.Register[0] < b.node.Register[0]
		},
		SwapFunc: func(a, b *cacheEntry, i, j int) {
			a.HeapIndex = i
			b.HeapIndex = j
		},
	}
	heap.Init(pq)

	head := utils.NewDLLNode(nil)
	tail := utils.NewDLLNode(nil)
	head.Next = tail
	tail.Prev = head

	var c = &optimizedCache{
		capacity:                   cacheCapacity,
		timestampsRegisterCapacity: timestampsRegisterCapacity,
		data:                       make(map[int]*cacheEntry, cacheCapacity),
		minHeap:                    pq,
		head:                       head,
		tail:                       tail,
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
	entry, ok := oc.data[key]
	if !ok {
		// if not exists and since its get operation, return nil
		return -1, errors.New("key does not exists in cache")
	}

	// record the new timestamp
	if err := recordAccess(entry.node, oc.timestamp); err != nil {
		return -1, err
	}

	// now since the timestamp is recorded, we need to check whether it should be kept in dll or
	// we should move it to heap
	if len(entry.node.Register) == cap(entry.node.Register) {
		// if the capacity is reached, and after reaching capacity we are accessing it first time, just
		// remove from dll otherwise no operation on dll to be performed
		if entry.DllNode != nil {
			dllNode := entry.DllNode
			dllNode.Remove()
			entry.DllNode = nil

			heap.Push(oc.minHeap, entry)
		} else {
			// push the node to heap
			heap.Fix(oc.minHeap, entry.HeapIndex)
		}
	} else {
		// if the capacity is not reached simple move the node to top as its the hot node for now
		// in dll
		dllNode := entry.DllNode
		dllNode.Remove()
		dllNode.Insert(oc.head)
	}

	return entry.node.Value, nil
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
	entry, ok := oc.data[key]
	if ok {
		// if an existing node, just update the value
		entry.node.Value = val

		// record the new access in the node
		if err := recordAccess(entry.node, oc.timestamp); err != nil {
			return -1, err
		}

		// if the capacity is reached, we should perform similar operation as GET (with
		// perspective to DLL and Heap)
		if len(entry.node.Register) == cap(entry.node.Register) {
			if entry.DllNode != nil {
				dllNode := entry.DllNode
				dllNode.Remove()
				entry.DllNode = nil

				heap.Push(oc.minHeap, entry)
			} else {
				// push the node to minheap
				heap.Fix(oc.minHeap, entry.HeapIndex)
			}
		} else {
			dllNode := entry.DllNode
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
		entry = &cacheEntry{}
		entry.node = cacheNode.NewNode(key, val, oc.timestampsRegisterCapacity)

		// record the new access in the node
		if err := recordAccess(entry.node, oc.timestamp); err != nil {
			return -1, err
		}

		// create a new dllNode and capture it in cache - because every new node will land in dll first
		dllNode := utils.NewDLLNode(entry)
		dllNode.Insert(oc.head)
		entry.DllNode = dllNode
	}

	// assign the key in the cache
	oc.data[key] = entry

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

		entry, ok := dllNode.Value.(*cacheEntry)
		if !ok || entry == nil {
			return errors.New("invalid dll node")
		}

		delete(oc.data, entry.node.Key)
		return nil
	}

	// if no node is present in dll then remove the node from heap as well as cache
	element := heap.Pop(oc.minHeap)
	pqItem := element.(*cacheEntry)

	// also delete the node from cache which got polled
	delete(oc.data, pqItem.node.Key)

	return nil
}

func recordAccess(n *cacheNode.Node, timestamp int64) error {
	if err := n.Validate(); err != nil {
		return err
	}

	if len(n.Register) == cap(n.Register) {
		n.Register = n.Register[1:]
	}

	n.Register = append(n.Register, timestamp)
	return nil
}
