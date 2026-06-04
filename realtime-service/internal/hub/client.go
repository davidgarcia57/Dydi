package hub

import (
	"context"
	"os"
	"strconv"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Client struct {
	groupID string
	userID  string
	conn    *websocket.Conn
	send    chan Event
}

func NewClient(groupID, userID string, conn *websocket.Conn) *Client {
	return &Client{
		groupID: groupID,
		userID:  userID,
		conn:    conn,
		send:    make(chan Event, 32),
	}
}

func pingInterval() time.Duration {
	s, _ := strconv.Atoi(os.Getenv("PING_INTERVAL_SECONDS"))
	if s <= 0 {
		s = 30
	}
	return time.Duration(s) * time.Second
}

func writeWait() time.Duration {
	s, _ := strconv.Atoi(os.Getenv("WRITE_WAIT_SECONDS"))
	if s <= 0 {
		s = 10
	}
	return time.Duration(s) * time.Second
}

// WritePump forwards hub events to the WebSocket connection.
// Keepalive pings prevent Render free-tier from dropping idle connections (~55s limit).
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval())
	defer ticker.Stop()

	for {
		select {
		case ev, ok := <-c.send:
			if !ok {
				c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeWait())
			err := wsjson.Write(writeCtx, c.conn, ev)
			cancel()
			if err != nil {
				return
			}

		// Keepalive: Render free-tier drops idle WS connections after ~55s
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait())
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// ReadPump reads and discards incoming messages from the client.
// This service is server-push only; the read loop exists to detect
// client disconnects via close frames or read errors.
// When it returns it cancels the shared context, which stops WritePump.
func (c *Client) ReadPump(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	for {
		if _, _, err := c.conn.Read(ctx); err != nil {
			return
		}
	}
}
