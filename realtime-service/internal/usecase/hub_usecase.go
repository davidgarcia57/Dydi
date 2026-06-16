package usecase

import (
	"log"
	"sync"
	"time"

	"github.com/dydi/realtime-service/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	// coldStartSeconds is the time from process start to the first WebSocket
	// connection — the paper's "WebSocket cold start" metric, made scrapable.
	coldStartSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "realtime_cold_start_seconds",
		Help: "Seconds from process start to the first client connection.",
	})
	// eventsDroppedTotal counts events dropped because a client's send buffer was
	// full (slow consumer) — feeds the paper's event-delivery-consistency metric.
	eventsDroppedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "realtime_events_dropped_total",
		Help: "Total events dropped due to a full client send buffer.",
	})
)

type HubUseCase struct {
	mu         sync.RWMutex
	rooms      map[string]map[*domain.Client]struct{}
	register   chan *domain.Client
	unregister chan *domain.Client
	broadcast  chan domain.Event

	startTime        time.Time
	firstConnectedAt time.Time
}

func NewHubUseCase() *HubUseCase {
	return &HubUseCase{
		rooms:      make(map[string]map[*domain.Client]struct{}),
		register:   make(chan *domain.Client),
		unregister: make(chan *domain.Client),
		broadcast:  make(chan domain.Event, 256),
		startTime:  time.Now(),
	}
}

func (h *HubUseCase) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.firstConnectedAt.IsZero() {
				h.firstConnectedAt = time.Now()
				cold := h.firstConnectedAt.Sub(h.startTime)
				coldStartSeconds.Set(cold.Seconds())
				log.Printf("realtime_cold_start_ms=%d", cold.Milliseconds())
			}
			if h.rooms[c.GroupID] == nil {
				h.rooms[c.GroupID] = make(map[*domain.Client]struct{})
			}
			h.rooms[c.GroupID][c] = struct{}{}
			h.mu.Unlock()
			connectionsTotal.Inc()

		case c := <-h.unregister:
			h.mu.Lock()
			removed := false
			if room, ok := h.rooms[c.GroupID]; ok {
				if _, exists := room[c]; exists {
					delete(room, c)
					if len(room) == 0 {
						delete(h.rooms, c.GroupID)
					}
					removed = true
				}
			}
			h.mu.Unlock()
			if removed {
				close(c.Send)
			}

		case ev := <-h.broadcast:
			eventsEmittedTotal.Inc()
			h.mu.RLock()
			for c := range h.rooms[ev.GroupID] {
				select {
				case c.Send <- ev:
					eventsDeliveredTotal.Inc()
				default:
					eventsDroppedTotal.Inc()
					go h.Unregister(c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Shutdown closes all active client connections cleanly.
// Called on SIGTERM to drain before the process exits.
func (h *HubUseCase) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for groupID, room := range h.rooms {
		for c := range room {
			select {
			case c.Send <- domain.Event{Type: domain.EventMemberOffline, GroupID: groupID, UserID: c.UserID}:
			default:
			}
			close(c.Send)
		}
	}
	h.rooms = make(map[string]map[*domain.Client]struct{})
}

func (h *HubUseCase) Register(c *domain.Client)   { h.register <- c }
func (h *HubUseCase) Unregister(c *domain.Client) { h.unregister <- c }

// Broadcast stamps the emit time (so clients can measure delivery latency) and
// enqueues the event for fan-out.
func (h *HubUseCase) Broadcast(ev domain.Event) {
	if ev.EmittedAt.IsZero() {
		ev.EmittedAt = time.Now()
	}
	h.broadcast <- ev
}

func (h *HubUseCase) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for _, room := range h.rooms {
		n += len(room)
	}
	return n
}

func (h *HubUseCase) RoomConnectionCount(groupID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.rooms[groupID]; ok {
		return len(room)
	}
	return 0
}
