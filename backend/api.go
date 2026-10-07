package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type VolcanoAlert struct {
	Name        string  `json:"name"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	AlertLevel  string  `json:"alert_level"`
	ColorCode   string  `json:"color_code"`
	Synopsis    string  `json:"synopsis"`
	Threat      string  `json:"threat"`
	Observatory string  `json:"observatory"`
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	q := r.URL.Query().Get("q")
	if q == "" {
		q = "sismo OR terremoto"
	}

	fp := gofeed.NewParser()
	feedURL := fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=es-419&gl=VE&ceid=VE:es-419", url.QueryEscape(q))

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch news"}`, http.StatusInternalServerError)
		return
	}

	type NewsItem struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		PubDate string `json:"pubDate"`
		Source  string `json:"source"`
	}

	var results []NewsItem
	limit := 15
	if len(feed.Items) < limit {
		limit = len(feed.Items)
	}

	for i := 0; i < limit; i++ {
		item := feed.Items[i]
		source := "Google News"
		if item.Extensions != nil && item.Extensions["source"] != nil {
			// Extract source if available
		}

		results = append(results, NewsItem{
			Title:   item.Title,
			Link:    item.Link,
			PubDate: item.Published,
			Source:  source,
		})
	}

	json.NewEncoder(w).Encode(results)
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "URL required", http.StatusBadRequest)
		return
	}

	if !allowedProxyDomain(targetURL) {
		http.Error(w, "Domain not allowed", http.StatusForbidden)
		return
	}

	isPlaylist := strings.Contains(targetURL, ".m3u8")

	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Proxy error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if isPlaylist {
		bodyBytes, _ := io.ReadAll(resp.Body)
		content := string(bodyBytes)

		baseURL, _ := url.Parse(targetURL)

		lines := strings.Split(content, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				var absoluteURL string
				if strings.HasPrefix(trimmed, "http") {
					absoluteURL = trimmed
				} else {
					ref, _ := url.Parse(trimmed)
					absoluteURL = baseURL.ResolveReference(ref).String()
				}
				lines[i] = fmt.Sprintf("/api/proxy/m3u8?url=%s", url.QueryEscape(absoluteURL))
			}
		}

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Write([]byte(strings.Join(lines, "\n")))
	} else {
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		io.Copy(w, resp.Body)
	}
}

func historicalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	targetDate := r.URL.Query().Get("date")
	startTime := r.URL.Query().Get("starttime")
	endTime := r.URL.Query().Get("endtime")

	if targetDate == "" && startTime == "" && endTime == "" {
		earthquakeMutex.RLock()
		defer earthquakeMutex.RUnlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":       len(cachedEarthquakes),
			"earthquakes": cachedEarthquakes,
		})
		return
	}

	// Determine start/end for USGS query
	var usgsStart, usgsEnd string
	var funvisisStart, funvisisEnd string

	if startTime != "" && endTime != "" {
		// Date range query
		usgsStart = startTime
		endParsed, _ := time.Parse("2006-01-02", endTime)
		usgsEnd = endParsed.Add(24 * time.Hour).Format("2006-01-02")
		funvisisStart = startTime
		funvisisEnd = endTime
	} else if startTime != "" {
		// From start date to now
		usgsStart = startTime
		usgsEnd = time.Now().UTC().Format("2006-01-02")
		funvisisStart = startTime
		funvisisEnd = usgsEnd
	} else {
		// Legacy single date query
		usgsStart = targetDate
		nextDayObj, _ := time.Parse("2006-01-02", targetDate)
		usgsEnd = nextDayObj.Add(24 * time.Hour).Format("2006-01-02")
		funvisisStart = targetDate
		funvisisEnd = targetDate
	}

	usgsHistUrl := fmt.Sprintf("https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson&starttime=%s&endtime=%s", usgsStart, usgsEnd)

	var histEarthquakes []Earthquake

	resp, err := http.Get(usgsHistUrl)
	if err == nil {
		defer resp.Body.Close()
		var usgsData struct {
			Features []struct {
				ID         string `json:"id"`
				Properties struct {
					Mag   float64     `json:"mag"`
					Place string      `json:"place"`
					Time  int64       `json:"time"`
					URL   string      `json:"url"`
					Felt  interface{} `json:"felt"`
					CDI   interface{} `json:"cdi"`
					MMI   interface{} `json:"mmi"`
					Alert string      `json:"alert"`
				} `json:"properties"`
				Geometry struct {
					Coordinates []float64 `json:"coordinates"` // lon, lat, depth
				} `json:"geometry"`
			} `json:"features"`
		}

		if json.NewDecoder(resp.Body).Decode(&usgsData) == nil {
			for _, f := range usgsData.Features {
				depth := "N/A"
				if len(f.Geometry.Coordinates) > 2 {
					depth = fmt.Sprintf("%.1f km", f.Geometry.Coordinates[2])
				}
				histEarthquakes = append(histEarthquakes, Earthquake{
					ID:          f.ID,
					Source:      "USGS",
					Magnitude:   f.Properties.Mag,
					Depth:       depth,
					Location:    f.Properties.Place,
					Time:        f.Properties.Time,
					Coordinates: []float64{f.Geometry.Coordinates[1], f.Geometry.Coordinates[0]},
					URL:         f.Properties.URL,
					Felt:        f.Properties.Felt,
					CDI:         f.Properties.CDI,
					MMI:         f.Properties.MMI,
					Alert:       f.Properties.Alert,
				})
			}
		}
	}

	var localHistFunvisis []Earthquake
	if funvisisStart == funvisisEnd {
		localHistFunvisis = getFunvisisHistoryByDate(funvisisStart)
	} else {
		localHistFunvisis = getFunvisisHistoryByDateRange(funvisisStart, funvisisEnd)
	}

	for _, f := range localHistFunvisis {
		duplicate := false
		for i, u := range histEarthquakes {
			timeDiff := math.Abs(float64(u.Time - f.Time))
			if timeDiff <= 300000 {
				dist := getDistanceFromLatLonInKm(u.Coordinates[0], u.Coordinates[1], f.Coordinates[0], f.Coordinates[1])
				if dist < 50 {
					histEarthquakes[i].Source = "USGS + FUNVISIS"
					duplicate = true
					break
				}
			}
		}
		if !duplicate {
			histEarthquakes = append(histEarthquakes, f)
		}
	}

	// Sort by time
	for i := 0; i < len(histEarthquakes); i++ {
		for j := i + 1; j < len(histEarthquakes); j++ {
			if histEarthquakes[i].Time < histEarthquakes[j].Time {
				histEarthquakes[i], histEarthquakes[j] = histEarthquakes[j], histEarthquakes[i]
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":        len(histEarthquakes),
		"earthquakes":  histEarthquakes,
		"isHistorical": true,
	})
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var sub PushSubscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid subscription payload", http.StatusBadRequest)
		return
	}

	if err := saveSubscription(sub); err != nil {
		http.Error(w, "Failed to save subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"subscribed"}`))
}

func vapidKeyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"publicKey": vapidPublicKey,
	})
}

func volcanoAlertsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	alerts := getVolcanoAlerts()
	if alerts == nil {
		alerts = []VolcanoAlert{}
	}
	json.NewEncoder(w).Encode(alerts)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	earthquakeMutex.RLock()
	count := len(cachedEarthquakes)
	earthquakeMutex.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cachedEarthquakes": count,
		"activeClients":     len(clients),
		"timestamp":         time.Now().Unix(),
	})
}
