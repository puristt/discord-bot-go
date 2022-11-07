/*
Package queue is basic queue implementation with slices.
Note that this implementation is not thread-safe because it fulfills the requirement of this project
*/
package queue

type items []interface{}

type Queue struct {
	items items
}

// New creates a new slice with given capacity(hint)
func New(hint int64) *Queue {
	return &Queue{
		items: make([]interface{}, 0, hint),
	}
}

// Len returns the number of items in this queue
func (q *Queue) Len() int64 {
	return int64(len(q.items))
}

// Empty returns true if this queue is empty. Otherwise, false.
func (q *Queue) Empty() bool {
	return len(q.items) == 0
}

// Enqueue appends given item to this queue
func (q *Queue) Enqueue(item interface{}) {
	q.items = append(q.items, item)
}

// Dequeue removes first item from this queue and returns removed item
func (q *Queue) Dequeue() interface{} {
	dItem := q.items[0]
	q.items = q.items[1:]
	return dItem
}
