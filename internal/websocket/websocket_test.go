package websocket

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}

	if hub.clients == nil {
		t.Error("hub.clients is nil")
	}

	if hub.broadcast == nil {
		t.Error("hub.broadcast channel is nil")
	}

	if hub.register == nil {
		t.Error("hub.register channel is nil")
	}

	if hub.unregister == nil {
		t.Error("hub.unregister channel is nil")
	}

	if hub.ctx == nil {
		t.Error("hub.ctx is nil")
	}

	if hub.cancel == nil {
		t.Error("hub.cancel is nil")
	}

	if hub.done == nil {
		t.Error("hub.done channel is nil")
	}

	// Start hub and clean up
	go hub.Run()
	time.Sleep(10 * time.Millisecond)
	hub.Shutdown()
}

func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)
	defer hub.Shutdown()

	// Initially should have 0 clients
	count := hub.ClientCount()
	if count != 0 {
		t.Errorf("ClientCount() = %d, want 0", count)
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Test broadcasting a message (should not block or panic)
	message := []byte("test message")
	hub.Broadcast(message)

	// Give it a moment to process
	time.Sleep(10 * time.Millisecond)

	// Test multiple broadcasts
	for i := 0; i < 5; i++ {
		hub.Broadcast([]byte("message"))
	}

	// Should not block or panic
}

func TestHub_BroadcastAfterShutdown(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Shutdown immediately
	hub.Shutdown()

	// Broadcasting after shutdown should not panic or block
	message := []byte("test message")
	hub.Broadcast(message)
}

func TestHub_Shutdown(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Give Run a moment to start
	time.Sleep(10 * time.Millisecond)

	// Shutdown should complete without error
	err := hub.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}
}

func TestHub_ShutdownWithTimeout(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a timeout channel
	done := make(chan error, 1)
	go func() {
		done <- hub.Shutdown()
	}()

	// Shutdown should complete within 1 second
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Shutdown() returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Shutdown() timed out")
	}
}

func TestHub_SendUpdate(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Test SendUpdate (should format and broadcast message)
	hub.SendUpdate("test_update", nil)

	// Give it a moment to process
	time.Sleep(10 * time.Millisecond)

	// Should not panic or block
}

func TestHub_SendUpdateWithData(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Test SendUpdate with data
	data := map[string]interface{}{
		"key": "value",
	}
	hub.SendUpdate("test_update", data)

	// Give it a moment to process
	time.Sleep(10 * time.Millisecond)

	// Should not panic or block
}

func TestHub_ConcurrentBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Test concurrent broadcasts from multiple goroutines
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				hub.Broadcast([]byte("message"))
				time.Sleep(time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or deadlock
}

