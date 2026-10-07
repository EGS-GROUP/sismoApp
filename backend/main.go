package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	db                 *sql.DB
	clients            = make(map[chan []byte]string)
	clientsMutex       sync.Mutex
	cachedEarthquakes  []Earthquake
	cachedAlerts       []TelegramAlert
	earthquakeMutex    sync.RWMutex
	alertMutex         sync.RWMutex
)

func main() {
	var err error
	// Initialize SQLite in WAL mode for high concurrency
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Error creating data directory: %v", err)
	}
	// If a previous (empty) directory is in the way of the DB file, back it up
	if info, statErr := os.Stat("data/sismos.db"); statErr == nil && info.IsDir() {
		backup := "data/sismos.db.bak"
		if _, existErr := os.Stat(backup); existErr == nil {
			backup = fmt.Sprintf("data/sismos.db.bak.%d", time.Now().Unix())
		}
		if err := os.Rename("data/sismos.db", backup); err != nil {
			log.Fatalf("Error backing up existing sismos.db directory: %v", err)
		}
		log.Printf("Backed up existing sismos.db directory to %s", backup)
	}
	db, err = sql.Open("sqlite", "file:data/sismos.db?cache=shared&_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatalf("Error opening db: %v", err)
	}
	defer db.Close()

	initDB(db)

	// Start background workers
	go fetchEarthquakeDataLoop()
	go fetchTelegramDataLoop()
	go fetchVolcanoAlertsLoop()

	// Rate limiter: 60 requests per minute per IP for general endpoints
	generalLimiter := newRateLimiter(60, time.Minute)
	go generalLimiter.cleanup()

	// Stricter limiter for push subscriptions: 10 per minute
	subscribeLimiter := newRateLimiter(10, time.Minute)
	go subscribeLimiter.cleanup()

	// API Routes with middleware
	http.HandleFunc("/api/stream", corsMiddleware(rateLimitMiddleware(generalLimiter, sseHandler)))
	http.HandleFunc("/api/news", corsMiddleware(rateLimitMiddleware(generalLimiter, newsHandler)))
	http.HandleFunc("/api/proxy/m3u8", corsMiddleware(rateLimitMiddleware(generalLimiter, proxyHandler)))
	http.HandleFunc("/api/earthquakes", corsMiddleware(rateLimitMiddleware(generalLimiter, historicalHandler)))
	http.HandleFunc("/api/subscribe", corsMiddleware(maxBodyMiddleware(4096, rateLimitMiddleware(subscribeLimiter, subscribeHandler))))
	http.HandleFunc("/api/vapidPublicKey", corsMiddleware(rateLimitMiddleware(generalLimiter, vapidKeyHandler)))
	http.HandleFunc("/api/volcano-alerts", corsMiddleware(rateLimitMiddleware(generalLimiter, volcanoAlertsHandler)))
	http.HandleFunc("/health", corsMiddleware(healthHandler))
	http.HandleFunc("/stats", corsMiddleware(rateLimitMiddleware(generalLimiter, statsHandler)))

	port := "8082"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	log.Printf("Servidor High-Performance (Go) corriendo en puerto %s...", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func getActiveCount() int {
	unique := make(map[string]bool)
	for _, id := range clients {
		if id != "" {
			unique[id] = true
		}
	}
	return len(unique)
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// CORS for development, adjust for prod
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("anon-%d", time.Now().UnixNano())
	}

	clientChan := make(chan []byte, 10)
	
	clientsMutex.Lock()
	clients[clientChan] = clientID
	activeCount := getActiveCount()
	clientsMutex.Unlock()
	
	// Broadcast active users immediately on connect
	go broadcastSSE("stats_update", map[string]int{"active_users": activeCount})

	defer func() {
		clientsMutex.Lock()
		delete(clients, clientChan)
		newCount := getActiveCount()
		clientsMutex.Unlock()
		close(clientChan)
		// Broadcast updated active users when someone disconnects
		go broadcastSSE("stats_update", map[string]int{"active_users": newCount})
	}()

	// Send initial state immediately
	earthquakeMutex.RLock()
	eqData, _ := json.Marshal(map[string]interface{}{"earthquakes": cachedEarthquakes})
	earthquakeMutex.RUnlock()

	alertMutex.RLock()
	alertData, _ := json.Marshal(map[string]interface{}{"alerts": cachedAlerts})
	alertMutex.RUnlock()

	volcanoData, _ := json.Marshal(map[string]interface{}{"volcano_alerts": getVolcanoAlerts()})

	fmt.Fprintf(w, "event: earthquakes_update\ndata: %s\n\n", eqData)
	fmt.Fprintf(w, "event: telegram_update\ndata: %s\n\n", alertData)
	fmt.Fprintf(w, "event: volcano_alerts_update\ndata: %s\n\n", volcanoData)
	flusher.Flush()

	// Listen for channel closure (client disconnect)
	notify := r.Context().Done()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, ":\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case msg := <-clientChan:
			if _, err := fmt.Fprintf(w, "%s\n\n", msg); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func broadcastSSE(eventName string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	msg := fmt.Sprintf("event: %s\ndata: %s", eventName, string(data))
	msgBytes := []byte(msg)

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for clientChan := range clients {
		// Non-blocking send
		select {
		case clientChan <- msgBytes:
		default:
			// If buffer is full, we can drop or log it. 
			// With SSE, we want to ensure clients don't block the server.
		}
	}
}
