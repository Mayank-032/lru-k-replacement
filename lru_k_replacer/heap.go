package lrukreplacer

type Item struct {
	Node *node
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Node.register[0] < pq[j].Node.register[0]
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Node.heapIndex = i
	pq[j].Node.heapIndex = j
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*Item)
	item.Node.heapIndex = pq.Len()
	*pq = append(*pq, item)

}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // Avoid memory leak
	*pq = old[0 : n-1]
	item.Node.heapIndex = -1
	return item
}
