package sse

import (
	"sync"
	"testing"
	"time"
)

func TestHubSubscribeAndBroadcast(t *testing.T) {
	hub := NewHub()
	groupID := "test-group-family-1"

	ch1, unsub1 := hub.Subscribe(groupID)
	defer unsub1()

	ch2, unsub2 := hub.Subscribe(groupID)
	defer unsub2()

	// Broadcast event
	testPayload := map[string]string{"message": "hallo keluarga"}
	hub.BroadcastToGroup(groupID, "TestEvent", testPayload)

	// Verify ch1 receives
	select {
	case evt := <-ch1:
		if evt.Event != "TestEvent" {
			t.Errorf("expected event TestEvent, got %s", evt.Event)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("ch1 timed out waiting for event")
	}

	// Verify ch2 receives
	select {
	case evt := <-ch2:
		if evt.Event != "TestEvent" {
			t.Errorf("expected event TestEvent, got %s", evt.Event)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("ch2 timed out waiting for event")
	}

	// Test Unsubscribe
	unsub1()

	// Broadcast again, only ch2 should receive
	hub.BroadcastToGroup(groupID, "SecondEvent", testPayload)

	select {
	case evt := <-ch2:
		if evt.Event != "SecondEvent" {
			t.Errorf("expected event SecondEvent, got %s", evt.Event)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("ch2 timed out waiting for second event")
	}

	// Verify ch1 is closed
	_, ok := <-ch1
	if ok {
		t.Error("expected ch1 to be closed after unsubscribe")
	}
}

func TestHubConcurrentBroadcast(t *testing.T) {
	hub := NewHub()
	groupID := "concurrent-group"

	ch, unsub := hub.Subscribe(groupID)
	defer unsub()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hub.BroadcastToGroup(groupID, "ConcurrentEvent", idx)
		}(i)
	}
	wg.Wait()

	received := 0
	for {
		select {
		case <-ch:
			received++
		default:
			goto done
		}
	}
done:
	if received == 0 {
		t.Fatal("expected to receive at least some concurrent events")
	}
}
