package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func websocketPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	serverConnections := make(chan *websocket.Conn, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverConnections <- conn
	}))
	t.Cleanup(server.Close)
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	serverConn := <-serverConnections
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})
	return serverConn, clientConn
}

func TestClientWritePump(t *testing.T) {
	serverConn, peer := websocketPair(t)
	client := &Client{Conn: serverConn, Send: make(chan []byte, 1)}
	done := make(chan struct{})
	go func() {
		client.WritePump()
		close(done)
	}()
	client.Send <- []byte("hello")
	messageType, message, err := peer.ReadMessage()
	if err != nil || messageType != websocket.TextMessage || string(message) != "hello" {
		t.Fatalf("ReadMessage() = %d, %q, %v", messageType, message, err)
	}
	close(client.Send)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("WritePump() did not stop after send channel closed")
	}
}

func TestClientReadPumpUnregistersAfterPeerClose(t *testing.T) {
	serverConn, peer := websocketPair(t)
	hub := NewHub()
	go hub.Run()
	client := &Client{UserID: 9, Conn: serverConn, Send: make(chan []byte, 1)}
	hub.Register(client)
	done := make(chan struct{})
	go func() {
		client.ReadPump(hub)
		close(done)
	}()
	if err := peer.WriteMessage(websocket.TextMessage, []byte("ignored inbound message")); err != nil {
		t.Fatal(err)
	}
	if err := peer.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ReadPump() did not stop after peer close")
	}
	select {
	case _, open := <-client.Send:
		if open {
			t.Fatal("unregistered client send channel is open")
		}
	case <-time.After(time.Second):
		t.Fatal("hub did not unregister read client")
	}
}
