/*
Package queue is basic queue implementation with slices.
Note that this implementation is not thread-safe because it fulfills the requirement of this project.
For simplicity, it also uses Song struct type instead of interface{} struct. If needed, it can be converted to interface.
*/
package queue

import (
	"github.com/puristt/discord-bot-go/model"
	"sync"
)

type items []model.Song

type Queue struct {
	items items
	lock  sync.RWMutex
}

// New creates a new slice with given capacity(hint)
func New(hint int64) *Queue {
	return &Queue{
		items: make(items, 0, hint),
	}
}

// Len returns the number of items in this queue
func (q *Queue) Len() int64 {
	q.lock.Lock()
	defer q.lock.Unlock()

	return int64(len(q.items))
}

// Empty returns true if this queue is empty. Otherwise, false.
func (q *Queue) Empty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.items) == 0
}

// Enqueue appends given item to this queue
func (q *Queue) Enqueue(item model.Song) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items = append(q.items, item)
}

// Dequeue removes first item from this queue and returns removed item
func (q *Queue) Dequeue() model.Song {
	q.lock.Lock()
	defer q.lock.Unlock()

	dItem := q.items[0]
	q.items = q.items[1:]
	return dItem
}

func (q *Queue) Front() model.Song {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.items[0] // TODO : if length == 0, will be thrown error
}

func (q *Queue) Dispose() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items = nil
}

func (q *Queue) PeekAll() []model.Song {
	q.lock.Lock()
	defer q.lock.Unlock()

	peekItems, ok := q.items.getAll()
	if !ok {
		return nil
	}
	return peekItems
}

func (items *items) getAll() ([]model.Song, bool) {
	length := len(*items)

	if length == 0 {
		return nil, false
	}
	return *items, true
}
