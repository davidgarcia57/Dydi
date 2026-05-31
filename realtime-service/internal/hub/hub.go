package hub

import "sync"

type Event struct {
	Type    string      `json:"type"`
	GroupID string      `json:"groupID"`
	UserID  string      `json:"userID"`
	Payload interface{} `json:"payload"`
}

type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
}

func New() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.rooms[c.groupID] == nil {
				h.rooms[c.groupID] = make(map[*Client]struct{})
			}
			h.rooms[c.groupID][c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			removed := false
			if room, ok := h.rooms[c.groupID]; ok {
				if _, exists := room[c]; exists {
					delete(room, c)
					if len(room) == 0 {
						delete(h.rooms, c.groupID)
					}
					removed = true
				}
			}
			h.mu.Unlock()
			if removed {
				close(c.send)
			}

		case ev := <-h.broadcast:
			h.mu.RLock()
			for c := range h.rooms[ev.GroupID] {
				select {
				case c.send <- ev:
				default:
					// Unregister client asynchronously to avoid concurrent map modification
					// under RLock. Unregister will close the channel safely.
					go h.Unregister(c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(c *Client)   { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }
func (h *Hub) Broadcast(ev Event)   { h.broadcast <- ev }

func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for _, room := range h.rooms {
		n += len(room)
	}
	return n
}

func (h *Hub) RoomConnectionCount(groupID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.rooms[groupID]; ok {
		return len(room)
	}
	return 0
}
