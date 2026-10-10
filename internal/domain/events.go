package domain

type EventBroadcaster interface {
	BroadcastToGroup(groupID string, eventName string, data any)
}

type SSEEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}
