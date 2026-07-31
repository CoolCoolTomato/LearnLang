package websocket

import (
	"bytes"
	"testing"
	"time"
)

func TestHubRegisterSendAndUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	first := &Client{UserID: 1, Send: make(chan []byte, 2)}
	second := &Client{UserID: 1, Send: make(chan []byte, 2)}
	other := &Client{UserID: 2, Send: make(chan []byte, 1)}
	hub.Register(first)
	hub.Register(second)
	hub.Register(other)
	hub.SendToUser(1, []byte("hello"))
	for _, client := range []*Client{first, second} {
		select {
		case got := <-client.Send:
			if !bytes.Equal(got, []byte("hello")) {
				t.Fatalf("message = %q", got)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for broadcast")
		}
	}
	select {
	case got := <-other.Send:
		t.Fatalf("other user received %q", got)
	default:
	}

	hub.Unregister(first)
	select {
	case _, open := <-first.Send:
		if open {
			t.Fatal("unregistered client channel remains open")
		}
	case <-time.After(time.Second):
		t.Fatal("unregister did not close send channel")
	}
	hub.SendToUser(1, []byte("again"))
	select {
	case got := <-second.Send:
		if string(got) != "again" {
			t.Fatalf("second message = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("remaining client did not receive message")
	}
}

func TestHubDropsMessageForFullClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	client := &Client{UserID: 1, Send: make(chan []byte, 1)}
	client.Send <- []byte("already full")
	hub.Register(client)
	done := make(chan struct{})
	go func() {
		hub.SendToUser(1, []byte("dropped"))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("broadcast blocked on a full client")
	}
}
