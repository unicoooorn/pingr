package repo

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/unicoooorn/pingr/internal/model"
)

func TestInMemory_SetGet(t *testing.T) {
	ctx := context.Background()
	im := NewInMemory()

	sub := "subsys1"
	want := model.Status("ok")

	if err := im.Set(ctx, sub, want); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	got, err := im.Get(ctx, sub)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Get returned %v want %v", got, want)
	}
}

func TestInMemory_GetNotFound(t *testing.T) {
	ctx := context.Background()
	im := NewInMemory()

	_, err := im.Get(ctx, "no-such-key")
	if err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestInMemory_Concurrent(t *testing.T) {
	ctx := context.Background()
	im := NewInMemory()

	var wg sync.WaitGroup
	writers := 50
	readers := 50
	ops_per_goroutine := 200

	for w := range make([]struct{}, writers) {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("k-%d", id%10)
			for range make([]struct{}, ops_per_goroutine) {
				if err := im.Set(ctx, key, model.Status("ok")); err != nil {
					t.Errorf("Set error: %v", err)
					return
				}
				time.Sleep(time.Microsecond)
			}
		}(w)
	}

	for r := range make([]struct{}, readers) {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("k-%d", id%10)
			for range make([]struct{}, ops_per_goroutine) {
				_, err := im.Get(ctx, key)
				if err != nil && err != ErrNotFound {
					t.Errorf("Get unexpected error: %v", err)
					return
				}
				time.Sleep(time.Microsecond)
			}
		}(r)
	}

	wg.Wait()
}
