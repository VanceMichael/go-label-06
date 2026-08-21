package stream

import (
	"context"
	"sync"

	"github.com/VanceMichael/go-base-airbridge/internal/domain"
)

type Bus[T any] struct {
	mu          sync.RWMutex
	subscribers map[int]*subscriber[T]
	next        int
}

func New[T any]() *Bus[T] { return &Bus[T]{subscribers: map[int]*subscriber[T]{}} }

func (b *Bus[T]) Subscribe(buffer int) (int, <-chan T, error) {
	if buffer < 1 {
		return 0, nil, domain.ErrInvalid
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.next
	b.next++
	state := newSubscriber[T](buffer)
	b.subscribers[id] = state
	return id, state.output, nil
}

func (b *Bus[T]) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if state, ok := b.subscribers[id]; ok {
		delete(b.subscribers, id)
		state.close()
	}
}

func (b *Bus[T]) Publish(ctx context.Context, v T) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, state := range b.subscribers {
		if err := state.enqueue(ctx, v); err != nil {
			return err
		}
	}
	return nil
}
