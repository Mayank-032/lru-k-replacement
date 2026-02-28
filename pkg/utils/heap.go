package utils

type PriorityQueue[T any] struct {
	items    []*T
	LessFunc func(a, b *T) bool      // caller provides comparison logic
	SwapFunc func(a, b *T, i, j int) // caller provides the swapping logic
}

func (pq *PriorityQueue[T]) Len() int {
	return len(pq.items)
}

func (pq *PriorityQueue[T]) Less(i, j int) bool {
	//.Node.register[0]
	// return pq.items[i] < pq.items[j]
	return pq.LessFunc(pq.items[i], pq.items[j])
}

func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.SwapFunc(pq.items[i], pq.items[j], i, j)
}

func (pq *PriorityQueue[T]) Push(x any) {
	items := x.(*T)
	// item.Node.heapIndex = pq.Len()
	pq.items = append(pq.items, items)
}

func (pq *PriorityQueue[T]) Pop() any {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[0 : n-1]
	return item
}
