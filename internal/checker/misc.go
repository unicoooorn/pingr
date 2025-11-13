package checker

import (
	"context"
	"fmt"
	"net/http"
	"net"
	"time"

	"github.com/go-ping/ping"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/unicoooorn/pingr/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func CheckHttpHealth(
	ctx context.Context,
	url string,
	headers map[string]string,
	timeout time.Duration,
) model.CheckResult {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return model.CheckResult{
			Status:  model.PingStatusNotOk,
			Details: err.Error(),
		}
	}
	defer resp.Body.Close()

	var status model.PingStatus
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = model.PingStatusOk
	} else {
		status = model.PingStatusNotOk
	}

	return model.CheckResult{
		Status:  status,
		Details: fmt.Sprintf("http status code: %d", resp.StatusCode),
	}
}

func CheckIcmpHealth(
	ctx context.Context,
	host string,
	timeout time.Duration,
) model.CheckResult {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	pinger.Count = 1
	pinger.Timeout = timeout

	err = pinger.Run()
	stats := pinger.Statistics()
	if err != nil || stats.PacketLoss > 0 {
		details := "No reply"
		if err != nil {
			details += " or error: " + err.Error()
		}
		return model.CheckResult{
			Status:  model.PingStatusNotOk,
			Details: details,
		}
	}
	return model.CheckResult{
		Status:  model.PingStatusOk,
		Details: stats.AvgRtt.String(),
	}
}

func CheckTcpHealth(ctx context.Context, host string, port int, timeout time.Duration) model.CheckResult {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	defer conn.Close()
	return model.CheckResult{
		Status: model.PingStatusOk,
		Details: "connected",
	}
}

func CheckRedisHealth(ctx context.Context, addr string, timeout time.Duration) model.CheckResult {
	opts := &redis.Options{
		Addr: addr,
		DialTimeout: timeout,
	}
	rdb := redis.NewClient(opts)
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	return model.CheckResult{
		Status: model.PingStatusOk,
		Details: "pong",
	}
}

func CheckPostgresHealth(ctx context.Context, dsn string, timeout time.Duration) model.CheckResult {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	defer db.Close()
	err = db.PingContext(ctx)
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	return model.CheckResult{
		Status: model.PingStatusOk,
		Details: "ok",
	}
}

func CheckGrpcHealth(ctx context.Context, addr string, timeout time.Duration) model.CheckResult {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	defer conn.Close()

	healthClient := healthpb.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		return model.CheckResult{Status: model.PingStatusNotOk, Details: err.Error()}
	}
	status := model.PingStatusOk
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		status = model.PingStatusNotOk
	}
	return model.CheckResult{
		Status:  status,
		Details: resp.GetStatus().String(),
	}
}
