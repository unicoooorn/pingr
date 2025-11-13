package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Конфигурация
var (
	spamPort      = "7777"
	targetCount   = 1000
	targetURL     = ""
	spamTriggered = make(chan struct{}, 1)
)

func main() {
	if len(os.Args) > 1 {
		targetURL = os.Args[1]
	}
	if len(os.Args) > 2 {
		if n, err := strconv.Atoi(os.Args[2]); err == nil && n > 0 {
			targetCount = n
		}
	}

	// Запускаем фоновый цикл 1 раз в секунду
	go func() {
		for {
			doBackgroundRequest()
			time.Sleep(20 * time.Second)
		}
	}()

	http.HandleFunc("/spam", handleSpam)
	log.Printf("Service started at :%s. Call http://localhost:%s/spam, чтобы начать спамить\n", spamPort, spamPort)
	log.Printf("Target spam URL is: %s", targetURL)
	log.Fatal(http.ListenAndServe(":"+spamPort, nil))
}

func handleSpam(w http.ResponseWriter, r *http.Request) {
	select {
	case spamTriggered <- struct{}{}:
		go spamRequests()
		log.Printf("/spam вызван — начинаю спамить запросами на %s (count=%d)", targetURL, targetCount)
		fmt.Fprintf(w, "Запуск спама на %s (count=%d)\n", targetURL, targetCount)
	default:
		log.Printf("/spam вызван — уже спамлю")
		fmt.Fprintln(w, "Спам уже запущен!")
	}
}

func spamRequests() {
	for i := 0; i < targetCount; i++ {
		resp, err := http.Get(targetURL)
		if err != nil {
			log.Printf("Request #%d failed: %v\n", i+1, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("Request #%d, status: %s, body: %s\n", i+1, resp.Status, string(body))
	}
}

func doBackgroundRequest() {
	resp, err := http.Get(targetURL)
	if err != nil {
		log.Printf("[BG] Background request failed: %v", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("[BG] Background request, status: %s, body: %s", resp.Status, string(body))
}