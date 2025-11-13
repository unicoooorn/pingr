package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Глобальные переменные для урлов и состояния
var (
	remoteURLs            []string // теперь тут урлы
	lastGetValueWasOK     bool
	lastGetValueMu        sync.RWMutex
	getValueDuration      = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "get_value_duration_seconds",
		Help: "Время ответа ручки /get_value",
	})
	getValueCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "get_value_requests_total",
		Help: "Счетчик запросов к ручке /get_value",
	})
	healthStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_health_status",
		Help: "Health status: 1=ok; 0=error",
	}, []string{"last_get_value_ok"})
	cpuUsageGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "app_cpu_usage_percent",
		Help: "CPU usage percent",
	})
	memUsageGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "app_mem_usage_percent",
		Help: "RAM usage percent",
	})
)

func init() {
	// Регистрируем метрики
	prometheus.MustRegister(getValueDuration, getValueCounter, healthStatus, cpuUsageGauge, memUsageGauge)
}

func main() {
	// Получаем список урлов из аргументов командной строки, через запятую
	if len(os.Args) > 1 {
		remoteURLs = strings.Split(os.Args[1], ",")
		for i, u := range remoteURLs {
			remoteURLs[i] = strings.TrimSpace(u)
		}
	} else {
		remoteURLs = []string{}
	}

	// Порт для нашего сервиса (и для prometheus теперь тоже)
	serverPort := "8080"
	if envPort := os.Getenv("APP_PORT"); envPort != "" {
		serverPort = envPort
	}

	// Стартуем сбор метрик CPU/RAM
	go collectSysMetrics()

	// HTTP хендлеры на одном сервере
	http.HandleFunc("/get_value", getValueHandler)
	http.HandleFunc("/health", healthHandler)
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Service (and Prometheus metrics) on :%s", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}

func getValueHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	getValueCounter.Inc()
	val, err := getSumOrRandom()
	duration := time.Since(start).Seconds()
	getValueDuration.Observe(duration)

	// Обновляем health статус
	lastGetValueMu.Lock()
	lastGetValueWasOK = (err == nil)
	lastGetValueMu.Unlock()

	if err == nil {
		healthStatus.WithLabelValues("1").Set(1)
		healthStatus.WithLabelValues("0").Set(0)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"value": val})
	} else {
		healthStatus.WithLabelValues("1").Set(0)
		healthStatus.WithLabelValues("0").Set(1)
		http.Error(w, "error: "+err.Error(), http.StatusServiceUnavailable)
	}
}

func getSumOrRandom() (int, error) {
	if len(remoteURLs) == 0 {
		return rand.Intn(1000), nil
	}
	sum := 0
	for _, url := range remoteURLs {
		resp, err := http.Get(url)
		if err != nil {
			return 0, fmt.Errorf("fail on url %s: %w", url, err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return 0, fmt.Errorf("bad body from url %s: %w", url, err)
		}
		var result struct {
			Value int `json:"value"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return 0, fmt.Errorf("bad json from url %s: %w", url, err)
		}
		sum += result.Value
	}
	return sum, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	lastGetValueMu.RLock()
	defer lastGetValueMu.RUnlock()
	if lastGetValueWasOK {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	} else {
		w.WriteHeader(500)
		w.Write([]byte("bad"))
	}
}

func collectSysMetrics() {
	for {
		if percent, err := cpu.Percent(0, false); err == nil && len(percent) > 0 {
			cpuUsageGauge.Set(percent[0])
		}
		if vm, err := mem.VirtualMemory(); err == nil {
			memUsageGauge.Set(vm.UsedPercent)
		}
		time.Sleep(2 * time.Second)
	}
}
