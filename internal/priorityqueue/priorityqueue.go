package priorityqueue

import (
	"sync"
	"time"

	"github.com/go-flutter-desktop/go-flutter/embedder"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value    embedder.FlutterTask // The value of the item
	FireTime time.Time            // The priority of the item in the queue.

	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue struct {
	queue []*Item
	sync.Mutex
}

// NewPriorityQueue create a new PriorityQueue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{}
	pq.queue = make([]*Item, 0)
	return pq
}

func (pq *PriorityQueue) Len() int { return len(pq.queue) }

func (pq *PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the lowest, not highest, priority so we use lower
	// than here.
	return pq.queue[i].FireTime.Before(pq.queue[j].FireTime)
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
	pq.queue[i].index = i
	pq.queue[j].index = j
}

// Push add a new priority/value pair in the queue. 0 priority = max.
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.queue)
	item := x.(*Item)
	item.index = n
	pq.queue = append(pq.queue, item)
}

// Pop Remove and return the highest item (lowest priority)
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.queue
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	pq.queue = old[0 : n-1]
	return item
}
