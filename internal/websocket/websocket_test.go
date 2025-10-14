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

// TestHub_BroadcastChannelFull tests broadcast when channel is full
func TestHub_BroadcastChannelFull(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Fill up the broadcast channel (capacity is 256)
	for i := 0; i < 300; i++ {
		hub.Broadcast([]byte("message"))
	}

	// Should not block or panic (extra messages should be dropped)
	time.Sleep(10 * time.Millisecond)
}

// TestHub_MultipleShutdown tests calling Shutdown multiple times
func TestHub_MultipleShutdown(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// First shutdown
	err := hub.Shutdown()
	if err != nil {
		t.Errorf("First Shutdown() returned error: %v", err)
	}

	// Second shutdown should also complete without error
	err = hub.Shutdown()
	if err != nil {
		t.Errorf("Second Shutdown() returned error: %v", err)
	}
}

// TestHub_ConcurrentClientCount tests concurrent access to ClientCount
func TestHub_ConcurrentClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Call ClientCount concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = hub.ClientCount()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestHub_SendUpdateMultipleTypes tests different update types
func TestHub_SendUpdateMultipleTypes(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	updateTypes := []string{
		"data_update",
		"connection",
		"disconnection",
		"error",
		"success",
	}

	for _, updateType := range updateTypes {
		hub.SendUpdate(updateType, nil)
		time.Sleep(time.Millisecond)
	}
}

// TestHub_BroadcastAndClientCountConcurrent tests broadcast and ClientCount together
func TestHub_BroadcastAndClientCountConcurrent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	done := make(chan bool, 2)

	// Goroutine 1: Broadcast messages
	go func() {
		for i := 0; i < 50; i++ {
			hub.Broadcast([]byte("test"))
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Check client count
	go func() {
		for i := 0; i < 50; i++ {
			_ = hub.ClientCount()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

// TestHub_SendUpdateAfterShutdown tests SendUpdate after shutdown
func TestHub_SendUpdateAfterShutdown(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	hub.Shutdown()

	// Should not panic or block
	hub.SendUpdate("test", nil)
}

// TestHub_InitialState tests the initial state of a new hub
func TestHub_InitialState(t *testing.T) {
	hub := NewHub()

	if hub.clients == nil {
		t.Error("clients map should be initialized")
	}

	if len(hub.clients) != 0 {
		t.Errorf("clients map should be empty, got %d clients", len(hub.clients))
	}

	if hub.ClientCount() != 0 {
		t.Errorf("ClientCount should be 0, got %d", hub.ClientCount())
	}
}

// TestHub_ContextCancellation tests context cancellation behavior
func TestHub_ContextCancellation(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Verify context is active
	select {
	case <-hub.ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Context is still active, as expected
	}

	// Shutdown and verify context is cancelled
	hub.Shutdown()

	select {
	case <-hub.ctx.Done():
		// Context is cancelled, as expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after shutdown")
	}
}

// BenchmarkHub_Broadcast benchmarks the Broadcast method
func BenchmarkHub_Broadcast(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	message := []byte("benchmark message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.Broadcast(message)
	}
}

// BenchmarkHub_ClientCount benchmarks the ClientCount method
func BenchmarkHub_ClientCount(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.ClientCount()
	}
}

// BenchmarkHub_SendUpdate benchmarks the SendUpdate method
func BenchmarkHub_SendUpdate(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.SendUpdate("benchmark", nil)
	}
}

