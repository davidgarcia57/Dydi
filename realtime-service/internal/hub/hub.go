package hub

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	EventCheckin        = "checkin"
	EventStreakUpdate   = "streak_update"
	EventMemberOnline   = "member_online"
	EventMemberOffline  = "member_offline"
	EventRouletteStart  = "roulette_start"
	EventRouletteResult = "roulette_result"
	EventDebtCreated    = "debt_created"
)

var (
	connectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "realtime_connections_total",
		Help: "Total WebSocket connections accepted.",
	})
	eventsEmittedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "realtime_events_emitted_total",
		Help: "Total events broadcast by the hub.",
	})
	eventsDeliveredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "realtime_events_delivered_total",
		Help: "Total events successfully written to client channels.",
	})
)

type Event struct {
	Type    string      `json:"type"`
	GroupID string      `json:"groupID"`
	UserID  string      `json:"userID"`
	Payload interface{} `json:"payload"`
}

type Hub struct {
	mu         sync.RWMutex
	rooms      map[string]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event

	startTime        time.Time
	firstConnectedAt time.Time
}

func New() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 256),
		startTime:  time.Now(),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.firstConnectedAt.IsZero() {
				h.firstConnectedAt = time.Now()
				log.Printf("realtime_cold_start_ms=%d", h.firstConnectedAt.Sub(h.startTime).Milliseconds())
			}
			if h.rooms[c.groupID] == nil {
				h.rooms[c.groupID] = make(map[*Client]struct{})
			}
			h.rooms[c.groupID][c] = struct{}{}
			h.mu.Unlock()
			connectionsTotal.Inc()

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
			eventsEmittedTotal.Inc()
			h.mu.RLock()
			for c := range h.rooms[ev.GroupID] {
				select {
				case c.send <- ev:
					eventsDeliveredTotal.Inc()
				default:
					go h.Unregister(c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Shutdown closes all active client connections cleanly.
// Called on SIGTERM to drain before the process exits.
func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for groupID, room := range h.rooms {
		for c := range room {
			select {
			case c.send <- Event{Type: EventMemberOffline, GroupID: groupID, UserID: c.userID}:
			default:
			}
			close(c.send)
		}
	}
	h.rooms = make(map[string]map[*Client]struct{})
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
