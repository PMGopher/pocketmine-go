package utils

import "container/heap"

type priority interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

type pqItem[TPriority priority, TValue any] struct {
	priority TPriority
	value    TValue
}

type pqHeap[TPriority priority, TValue any] []pqItem[TPriority, TValue]

func (h pqHeap[TPriority, TValue]) Len() int { return len(h) }

// Less is reversed relative to a normal min-heap-by-priority: this mirrors
// ReversePriorityQueue::compare(), which negates SplPriorityQueue's default max-heap
// ordering so that the lowest priority value is extracted first.
func (h pqHeap[TPriority, TValue]) Less(i, j int) bool { return h[i].priority < h[j].priority }
func (h pqHeap[TPriority, TValue]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *pqHeap[TPriority, TValue]) Push(x any) {
	*h = append(*h, x.(pqItem[TPriority, TValue]))
}

func (h *pqHeap[TPriority, TValue]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// ReversePriorityQueue is a port of pocketmine\utils\ReversePriorityQueue: a priority queue
// that extracts the lowest priority value first (the reverse of SplPriorityQueue's default).
type ReversePriorityQueue[TPriority priority, TValue any] struct {
	h pqHeap[TPriority, TValue]
}

func NewReversePriorityQueue[TPriority priority, TValue any]() *ReversePriorityQueue[TPriority, TValue] {
	return &ReversePriorityQueue[TPriority, TValue]{}
}

func (q *ReversePriorityQueue[TPriority, TValue]) Insert(value TValue, priority TPriority) {
	heap.Push(&q.h, pqItem[TPriority, TValue]{priority: priority, value: value})
}

func (q *ReversePriorityQueue[TPriority, TValue]) Extract() TValue {
	return heap.Pop(&q.h).(pqItem[TPriority, TValue]).value
}

func (q *ReversePriorityQueue[TPriority, TValue]) Count() int {
	return len(q.h)
}

func (q *ReversePriorityQueue[TPriority, TValue]) IsEmpty() bool {
	return len(q.h) == 0
}

// Current mirrors SplPriorityQueue::current(): returns the next value to be Extract()ed, without
// removing it. container/heap always keeps the minimum element at index 0, so this is O(1).
func (q *ReversePriorityQueue[TPriority, TValue]) Current() TValue {
	return q.h[0].value
}
