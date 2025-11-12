package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
)

func TestHttpChecker_Local200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, test!"))
	}))
	defer ts.Close()

	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"localtest": {
				Type:    "http",
				URL:     ts.URL,
				Timeout: 2,
				Headers: nil,
			},
		},
	}

	c := NewHttpChecker(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := c.Check(ctx, "localtest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != model.PingStatusOk {
		t.Errorf("expected OK status, got %v: %s", res.Status, res.Details)
	}
	if want := "http status code: 200"; res.Details != want {
		t.Errorf("unexpected details: got %q, want %q", res.Details, want)
	}
}
