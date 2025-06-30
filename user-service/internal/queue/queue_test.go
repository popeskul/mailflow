package queue_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/popeskul/mailflow/user-service/internal/domain"
	"github.com/popeskul/mailflow/user-service/internal/queue"
)

func TestEmailQueue_EnqueueAndSize(t *testing.T) {
	logger := zap.NewNop()
	q := queue.NewEmailQueue(10, logger)

	// Test empty queue
	if size := q.Size(); size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}

	// Test enqueue
	email := &domain.Email{
		ID:      "test-1",
		To:      "test@example.com",
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	err := q.Enqueue(email)
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	// Test size
	if size := q.Size(); size != 1 {
		t.Errorf("Expected size 1, got %d", size)
	}
}

func TestEmailQueue_QueueFull(t *testing.T) {
	logger := zap.NewNop()
	q := queue.NewEmailQueue(2, logger) // Small buffer

	// Fill the queue
	for i := 0; i < 2; i++ {
		email := &domain.Email{
			ID:      "test-" + string(rune(i)),
			To:      "test@example.com",
			Subject: "Test Subject",
			Body:    "Test Body",
		}
		err := q.Enqueue(email)
		if err != nil {
			t.Fatalf("Failed to enqueue item %d: %v", i, err)
		}
	}

	// Try to enqueue when full (should not block, will return error)
	email := &domain.Email{
		ID:      "overflow",
		To:      "test@example.com",
		Subject: "Overflow",
		Body:    "Should fail",
	}

	err := q.Enqueue(email)
	if err == nil {
		t.Error("Expected error when queue is full, got nil")
	}
}

func TestEmailQueue_StartStop(t *testing.T) {
	logger := zap.NewNop()
	q := queue.NewEmailQueue(10, logger)

	processed := make(chan *domain.Email, 1)

	processor := func(email *domain.Email) error {
		processed <- email
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start processing
	go q.Start(ctx, processor)

	// Enqueue an email
	email := &domain.Email{
		ID:      "test-start-stop",
		To:      "test@example.com",
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	err := q.Enqueue(email)
	if err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	// Wait for processing
	select {
	case processedEmail := <-processed:
		if processedEmail.ID != email.ID {
			t.Errorf("Expected email ID %s, got %s", email.ID, processedEmail.ID)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Email was not processed within timeout")
	}

	// Stop the queue
	q.Stop()
}
