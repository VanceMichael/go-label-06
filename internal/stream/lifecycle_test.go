package stream

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestUnsubscribeReleasesPublisherBlockedBySlowConsumer(t *testing.T) {
	bus := New[int]()
	id, output, err := bus.Subscribe(1)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	if err := bus.Publish(context.Background(), 1); err != nil {
		t.Fatalf("publish first value: %v", err)
	}
	waitForStreamState(t, func() bool { return len(output) == 1 })
	if err := bus.Publish(context.Background(), 2); err != nil {
		t.Fatalf("publish second value: %v", err)
	}
	if err := bus.Publish(context.Background(), 3); err != nil {
		t.Fatalf("publish third value: %v", err)
	}
	waitForStreamState(t, func() bool { return len(bus.subscribers[id].inbox) == 1 })

	publishCtx, cancelPublish := context.WithCancel(context.Background())
	publishResult := make(chan error, 1)
	go func() {
		publishResult <- bus.Publish(publishCtx, 4)
	}()
	select {
	case err := <-publishResult:
		t.Fatalf("fourth publish did not wait for slow consumer: %v", err)
	case <-time.After(25 * time.Millisecond):
	}

	unsubscribed := make(chan struct{})
	go func() {
		bus.Unsubscribe(id)
		close(unsubscribed)
	}()
	select {
	case <-unsubscribed:
		cancelPublish()
		if err := <-publishResult; err != nil {
			t.Fatalf("publisher returned after unsubscribe: %v", err)
		}
		return
	case <-time.After(150 * time.Millisecond):
	}

	cancelPublish()
	if err := <-publishResult; !errors.Is(err, context.Canceled) {
		t.Fatalf("publisher cleanup error = %v", err)
	}
	select {
	case <-unsubscribed:
	case <-time.After(time.Second):
		t.Fatal("unsubscribe remained blocked during cleanup")
	}
	t.Fatal("unsubscribe blocked behind a stalled publisher")
}

func waitForStreamState(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !ready() {
		if time.Now().After(deadline) {
			t.Fatal("stream did not reach expected backpressure state")
		}
		time.Sleep(time.Millisecond)
	}
}
