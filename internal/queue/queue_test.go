package queue

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryQueueEnqueueDequeue(t *testing.T) {
	q := NewInMemoryQueue()
	job := Job{ID: "run-1", Type: JobTypeDeploy, CreatedAt: time.Now().UTC()}
	if err := q.Enqueue(context.Background(), job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	depth, _ := q.Depth(context.Background(), JobTypeDeploy)
	if depth != 1 {
		t.Fatalf("expected depth 1, got %d", depth)
	}
	out, err := q.Dequeue(context.Background(), []JobType{JobTypeDeploy}, time.Second)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if out.ID != "run-1" {
		t.Fatalf("unexpected job id %q", out.ID)
	}
}

func TestInMemoryQueueTimeout(t *testing.T) {
	q := NewInMemoryQueue()
	_, err := q.Dequeue(context.Background(), []JobType{JobTypeDeploy}, 10*time.Millisecond)
	if err != ErrEmpty {
		t.Fatalf("expected ErrEmpty, got %v", err)
	}
}
