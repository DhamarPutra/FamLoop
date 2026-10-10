package sse

import (
	"sync"

	"famloop/internal/domain"
)

type Hub struct {
	mu     sync.RWMutex
	groups map[string]map[chan domain.SSEEvent]bool
}

func NewHub() *Hub {
	return &Hub{
		groups: make(map[string]map[chan domain.SSEEvent]bool),
	}
}

// Subscribe mendaftarkan client ke grup tertentu dan mengembalikan channel event serta fungsi unsubscribe
func (h *Hub) Subscribe(groupID string) (chan domain.SSEEvent, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan domain.SSEEvent, 64)
	if h.groups[groupID] == nil {
		h.groups[groupID] = make(map[chan domain.SSEEvent]bool)
	}
	h.groups[groupID][ch] = true

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if groupSubs, exists := h.groups[groupID]; exists {
			delete(groupSubs, ch)
			close(ch)
			if len(groupSubs) == 0 {
				delete(h.groups, groupID)
			}
		}
	}

	return ch, unsubscribe
}

// BroadcastToGroup menyiarkan event ke seluruh client yang sedang terhubung di grup tersebut (non-blocking)
func (h *Hub) BroadcastToGroup(groupID string, eventName string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	groupSubs, exists := h.groups[groupID]
	if !exists || len(groupSubs) == 0 {
		return
	}

	evt := domain.SSEEvent{
		Event: eventName,
		Data:  data,
	}

	for ch := range groupSubs {
		select {
		case ch <- evt:
		default:
			// Client channel penuh atau lag, drop event untuk menjaga kestabilan goroutine
		}
	}
}
