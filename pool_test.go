package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// dialTestWS opens a real client WS conn to a silent local server.
func dialTestWS(t *testing.T) (*websocket.Conn, func()) {
	t.Helper()
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_, _, _ = c.ReadMessage() // block until client closes
	}))
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		srv.Close()
		t.Fatal(err)
	}
	return conn, func() { _ = conn.Close(); srv.Close() }
}

func TestPoolDiscardsAgedConn(t *testing.T) {
	conn, cleanup := dialTestWS(t)
	defer cleanup()

	p := newWSPool()
	key := dcKey{DC: 1}
	p.idle[key] = []pooledWS{{Conn: conn, Created: time.Now().Add(-2 * wsPoolMaxAge)}}

	cfg := &Config{PoolSize: 0} // PoolSize 0 -> no background refill
	if got := p.get(cfg, key, "1.2.3.4", []string{"d"}); got != nil {
		t.Error("aged pooled conn must be discarded (get -> nil)")
	}
}

func TestPoolReturnsFreshConn(t *testing.T) {
	conn, cleanup := dialTestWS(t)
	defer cleanup()

	p := newWSPool()
	key := dcKey{DC: 1}
	p.idle[key] = []pooledWS{{Conn: conn, Created: time.Now()}}

	cfg := &Config{PoolSize: 0}
	if got := p.get(cfg, key, "1.2.3.4", []string{"d"}); got != conn {
		t.Error("fresh pooled conn must be returned as-is")
	}
}
