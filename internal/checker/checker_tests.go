package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"net"

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

	c := NewChecker(&cfg)
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

func TestCheckerImpl_TcpOK(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().(*net.TCPAddr)

	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"mytcp": {
				Type:    "tcp",
				Host:    addr.IP.String(),
				Port:    addr.Port,
				Timeout: 2,
			},
		},
	}
	c := NewChecker(&cfg)
	ctx := context.Background()

	res, err := c.Check(ctx, "mytcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != model.PingStatusOk {
		t.Errorf("expected OK, got %v", res.Status)
	}
}

func TestCheckerImpl_Icmp_Localhost(t *testing.T) {
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"mylocal": {
				Type:    "icmp",
				Host:    "127.0.0.1",
				Timeout: 2,
			},
		},
	}
	c := NewChecker(&cfg)
	ctx := context.Background()

	res, err := c.Check(ctx, "mylocal")
	if err != nil {
		t.Skip("ICMP may be blocked in CI environments")
	}
	if res.Status != model.PingStatusOk {
		t.Errorf("expected OK, got %v", res.Status)
	}
}

func TestCheckerImpl_Redis_Unreachable(t *testing.T) {
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"myredis": {
				Type:    "redis",
				Host:    "127.0.0.1",
				Port:    65000, // скорее всего свободно
				Timeout: 1,
			},
		},
	}
	c := NewChecker(&cfg)
	ctx := context.Background()

	res, err := c.Check(ctx, "myredis")
	if err == nil {
		t.Error("expected error for unavailable redis, got nil")
	}
	if res.Status != model.PingStatusNotOk {
		t.Errorf("expected NotOk, got %v", res.Status)
	}
}

func TestCheckerImpl_Postgres_InvalidDSN(t *testing.T) {
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"mypg": {
				Type:    "postgres",
				URL:     "postgres://fake:fake@localhost:65000/db",
				Timeout: 1,
			},
		},
	}
	c := NewChecker(&cfg)
	ctx := context.Background()

	res, err := c.Check(ctx, "mypg")
	if err == nil {
		t.Error("expected error for invalid postgres dsn, got nil")
	}
	if res.Status != model.PingStatusNotOk {
		t.Errorf("expected NotOk, got %v", res.Status)
	}
}
