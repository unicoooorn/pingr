package sender

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Test successful POST to Telegram API: handler decodes JSON and returns 200.
func TestSender_SendAlert_Success(t *testing.T) {
	// Prepare test server to simulate Telegram API
	var received struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/bottoken/sendMessage") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(b, &received); err != nil {
			t.Fatalf("unmarshal: %v; body: %s", err, string(b))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	api := NewTgApi(server.URL, "token", "chat123")
	ctx := context.Background()
	msg := "hello world"
	if err := api.SendAlert(ctx, msg); err != nil {
		t.Fatalf("SendAlert returned error: %v", err)
	}

	if received.ChatID != "chat123" {
		t.Fatalf("unexpected chat_id: %q", received.ChatID)
	}
	if received.Text != msg {
		t.Fatalf("unexpected text: %q", received.Text)
	}
}

// Test non-200 response handling: expect error with status code included.
func TestSender_SendAlert_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	api := NewTgApi(server.URL, "token", "chat123")
	ctx := context.Background()
	if err := api.SendAlert(ctx, "x"); err == nil {
		t.Fatalf("expected error for non-200 response, got nil")
	} else {
		if !strings.Contains(err.Error(), "tg api returned status") {
			t.Fatalf("unexpected error message: %v", err)
		}
	}
}

// Test context cancellation: server sleeps, context times out -> expect wrapped context error.
func TestSender_SendAlert_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	api := NewTgApi(server.URL, "token", "chat123")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := api.SendAlert(ctx, "x")
	if err == nil {
		t.Fatalf("expected error due to context timeout, got nil")
	}
	
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got: %v", err)
	}
}

func contains(slice []string, v string) bool {
	for _, x := range slice {
		if x == v {
			return true
		}
	}
	return false
}

func TestCollectFromUpdates_StartStop(t *testing.T) {
	updates := make(chan tgbotapi.Update)

	go func() {
		u1 := tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1001}, Text: "/start"}}
		updates <- u1
		updates <- u1

		u2 := tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 2002}, Text: "/stop"}}
		updates <- u2

		u3 := tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 3003}, Text: "hello"}}
		updates <- u3

		close(updates)
	}()

	ctx := context.Background()
	starts, stops, err := collectFromUpdates(ctx, updates)
	if err == nil || err.Error() != "updates channel closed" {
		t.Fatalf("expected updates channel closed error, got %v", err)
	}

	if !contains(starts, "1001") {
		t.Fatalf("starts missing 1001: %#v", starts)
	}
	if !contains(stops, "2002") {
		t.Fatalf("stops missing 2002: %#v", stops)
	}
	if contains(starts, "3003") || contains(stops, "3003") {
		t.Fatalf("non-command chat unexpectedly present")
	}
}

func TestCollectFromUpdates_ContextCancel(t *testing.T) {
	updates := make(chan tgbotapi.Update)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	var starts, stops []string
	var err error
	go func() {
		starts, stops, err = collectFromUpdates(ctx, updates)
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	<-done

	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context cancel error, got %v", err)
	}
	if len(starts) != 0 || len(stops) != 0 {
		t.Fatalf("expected no starts/stops on cancel, got starts=%v stops=%v", starts, stops)
	}
}