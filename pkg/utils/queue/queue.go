/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package queue

import (
	"container/list"
	"context"
	"sync"
)

// Queue is a generic queue interface.
type Queue[T any] interface {
	// Add adds an item to the queue.
	Add(item T)
	// Get returns an item from the queue.
	Get() (T, bool)
	// GetOrWait returns an item from the queue or waits until an item is added.
	GetOrWait() T
	// GetOrWaitWithContext returns an item from the queue or waits until an item is added or the context is done.
	GetOrWaitWithContext(ctx context.Context) (T, bool)
	// Len returns the number of items in the queue.
	Len() int
}

// queue is a generic Queue implementation.
type queue[T any] struct {
	base *list.List

	signal chan struct{}
	mut    sync.Mutex
}

// NewQueue returns a new Queue.
func NewQueue[T any]() Queue[T] {
	return &queue[T]{
		base:   list.New(),
		signal: make(chan struct{}, 1),
	}
}

func (q *queue[T]) Add(item T) {
	q.mut.Lock()
	defer q.mut.Unlock()

	q.base.PushBack(item)

	// Signal that an item was added.
	if len(q.signal) == 0 {
		select {
		case q.signal <- struct{}{}:
		default:
		}
	}
}

func (q *queue[T]) Get() (t T, ok bool) {
	q.mut.Lock()
	defer q.mut.Unlock()
	item := q.base.Front()
	if item == nil {
		return t, false
	}
	q.base.Remove(item)
	return item.Value.(T), true
}

func (q *queue[T]) GetOrWait() T {
	t, ok := q.Get()
	if ok {
		return t
	}

	// Wait for an item to be added.
	for range q.signal {
		t, ok = q.Get()
		if ok {
			return t
		}
	}
	panic("unreachable")
}

func (q *queue[T]) GetOrWaitWithContext(ctx context.Context) (T, bool) {
	t, ok := q.Get()
	if ok {
		return t, ok
	}

	// Wait for an item to be added or the context to be done.
	for {
		select {
		case <-ctx.Done():
			return t, false
		case <-q.signal:
			t, ok = q.Get()
			if ok {
				return t, ok
			}
		}
	}
}

func (q *queue[T]) Len() int {
	q.mut.Lock()
	defer q.mut.Unlock()
	return q.base.Len()
}
