package utils

type DllNode struct {
	Value interface{}
	Prev  *DllNode
	Next  *DllNode
}

func NewDLLNode(val interface{}) *DllNode {
	return &DllNode{
		Value: val,
	}
}

func (dll *DllNode) Insert(node1 *DllNode) {
	nextNode := node1.Next

	// setting first pointer
	node1.Next = dll
	dll.Prev = node1

	// setting nextNode
	dll.Next = nextNode
	nextNode.Prev = dll
}

func (dll *DllNode) Remove() {
	prevNode := dll.Prev
	nextNode := dll.Next

	prevNode.Next = nextNode
	nextNode.Prev = prevNode
}
