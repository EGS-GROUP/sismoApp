package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Earthquake struct {
	ID          string      `json:"id"`
	Source      string      `json:"source"`
	Magnitude   float64     `json:"magnitude"`
	Depth       string      `json:"depth"`
	Location    string      `json:"location"`
	Time        int64       `json:"time"`
	Coordinates []float64   `json:"coordinates"` // [lat, lon]
	URL         string      `json:"url"`
	Felt        interface{} `json:"felt"`
	CDI         interface{} `json:"cdi"`
	MMI         interface{} `json:"mmi"`
	Alert       string      `json:"alert"`
}

type TelegramAlert struct {
	Group    string `json:"group"`
	Link     string `json:"link"`
	Text     string `json:"text"`
	FullText string `json:"fullText"`
	Type     string `json:"type"`
	Image    string `json:"image"`
	Video    string `json:"video"`
	Date     string `json:"date"`
}

var telegramChannels = []struct{ ID, Name string }{
	{"monitornewsve", "@monitornewsve"},
	{"angelreporta2020", "@angelreporta2020"},
	{"Reportesucesos18", "@Reportesucesos18"},
	{"PatriaDigital", "@PatriaDigital"},
	{"Alertas24", "@Alertas24"},
	{"noticiasencalienteupata", "@noticiasencalienteupata"},
	{"davidglockvzla", "@davidglockvzla"},
	{"noticiasvam", "@noticiasvam"},
	{"ONSAveDMO", "@ONSAveDMO"},
}

// Haversine formula
func getDistanceFromLatLonInKm(lat1, lon1, lat2, lon2 float64) float64 {
	var R float64 = 6371
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func fetchEarthquakeDataLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Run once immediately
	fetchEarthquakeData()

	for range ticker.C {
		fetchEarthquakeData()
	}
}

func fetchEarthquakeData() {
	usgsURL := "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_week.geojson"
	emscURL := "https://www.seismicportal.eu/fdsnws/event/1/query?limit=50&format=json"
	funvisisURL := "http://www.funvisis.gob.ve/maravilla.json"

	var allScraped []Earthquake

	client := &http.Client{Timeout: 15 * time.Second}

	// 1. Fetch USGS
	if resp, err := client.Get(usgsURL); err == nil {
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
		if err := json.NewDecoder(resp.Body).Decode(&usgsData); err == nil {
			for _, f := range usgsData.Features {
				depth := "N/A"
				if len(f.Geometry.Coordinates) > 2 {
					depth = fmt.Sprintf("%.1f km", f.Geometry.Coordinates[2])
				}
				allScraped = append(allScraped, Earthquake{
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

	// 2. Fetch EMSC
	if resp, err := client.Get(emscURL); err == nil {
		defer resp.Body.Close()
		var emscData struct {
			Features []struct {
				ID         string `json:"id"`
				Properties struct {
					Mag         float64 `json:"mag"`
					FlynnRegion string  `json:"flynn_region"`
					Time        string  `json:"time"`
					Depth       float64 `json:"depth"`
					Unid        string  `json:"unid"`
				} `json:"properties"`
				Geometry struct {
					Coordinates []float64 `json:"coordinates"` // lon, lat, depth
				} `json:"geometry"`
			} `json:"features"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&emscData); err == nil {
			for _, f := range emscData.Features {
				depth := fmt.Sprintf("%.1f km", math.Abs(f.Properties.Depth))
				t, err := time.Parse(time.RFC3339Nano, f.Properties.Time)
				if err != nil {
					t = time.Now()
				}
				allScraped = append(allScraped, Earthquake{
					ID:          "emsc-" + f.Properties.Unid,
					Source:      "EMSC",
					Magnitude:   f.Properties.Mag,
					Depth:       depth,
					Location:    f.Properties.FlynnRegion,
					Time:        t.UnixMilli(),
					Coordinates: []float64{f.Geometry.Coordinates[1], f.Geometry.Coordinates[0]},
					URL:         "https://www.emsc-csem.org/Earthquake_information/earthquake.php?id=" + f.Properties.Unid,
				})
			}
		}
	}

	// 3. Fetch FUNVISIS
	if resp, err := client.Get(funvisisURL); err == nil {
		defer resp.Body.Close()
		var funvData struct {
			Features []struct {
				Properties struct {
					PostalCode string `json:"postalCode"` // Date
					City       string `json:"city"`       // Time
					Lat        string `json:"lat"`
					Long       string `json:"long"`
					Phone      string `json:"phone"` // Mag
					State      string `json:"state"` // Depth
					Address    string `json:"address"`
				} `json:"properties"`
			} `json:"features"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&funvData); err == nil {
			for _, f := range funvData.Features {
				p := f.Properties

				cityTime := p.City
				if len(cityTime) == 4 {
					cityTime = "0" + cityTime
				}

				var t time.Time
				layout := "02-01-2006T15:04:00-07:00"
				dtStr := fmt.Sprintf("%sT%s:00-04:00", p.PostalCode, cityTime)
				t, err := time.Parse(layout, dtStr)
				if err != nil {
					log.Printf("FUNVISIS: error parseando fecha '%s': %v", dtStr, err)
					continue
				}

				// Rechazar timestamps en el futuro (protección contra datos corruptos)
				if t.After(time.Now().Add(1 * time.Minute)) {
					log.Printf("FUNVISIS: descartado sismo con timestamp futuro: %s (%s)", p.Address, dtStr)
					continue
				}

				lat, _ := strconv.ParseFloat(p.Lat, 64)
				lon, _ := strconv.ParseFloat(p.Long, 64)
				mag, _ := strconv.ParseFloat(p.Phone, 64)

				id := fmt.Sprintf("funvisis-%d-%.4f-%.4f", t.UnixMilli(), lat, lon)

				loc := p.Address
				if !strings.Contains(strings.ToLower(loc), "venezuela") {
					loc += ", Venezuela"
				}

				eq := Earthquake{
					ID:          strings.ReplaceAll(id, ".", "_"),
					Source:      "FUNVISIS",
					Magnitude:   mag,
					Depth:       p.State,
					Location:    loc,
					Time:        t.UnixMilli(),
					Coordinates: []float64{lat, lon},
					URL:         "http://www.funvisis.gob.ve/",
					Felt:        "Encuesta Disponible",
					Alert:       "Boletín Oficial",
				}

				saveFunvisisHistory(eq)
			}
		}
	}

	// 4. Fetch CSN (Chile)
	csnURL := "https://api.xor.cl/sismo/recent"
	if resp, err := client.Get(csnURL); err == nil {
		defer resp.Body.Close()
		var csnData struct {
			Events []struct {
				ID        string  `json:"id"`
				UTCDate   string  `json:"utc_date"` // "2026-06-28 15:56:15"
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
				Depth     float64 `json:"depth"`
				Magnitude struct {
					Value float64 `json:"value"`
				} `json:"magnitude"`
				GeoRef string `json:"geo_reference"`
				URL    string `json:"url"`
			} `json:"events"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&csnData); err == nil {
			for _, f := range csnData.Events {
				depth := fmt.Sprintf("%.1f km", f.Depth)
				t, err := time.Parse("2006-01-02 15:04:05", f.UTCDate)
				if err != nil {
					t = time.Now()
				}
				allScraped = append(allScraped, Earthquake{
					ID:          "csn-" + f.ID,
					Source:      "CSN",
					Magnitude:   f.Magnitude.Value,
					Depth:       depth,
					Location:    f.GeoRef + ", Chile",
					Time:        t.UnixMilli(),
					Coordinates: []float64{f.Latitude, f.Longitude},
					URL:         f.URL,
				})
			}
			log.Printf("CSN events parsed: %d", len(csnData.Events))
		} else {
			log.Printf("Error decoding CSN: %v", err)
		}
	}

	// 5. Fetch PRSN
	prsnURL := "http://www.prsn.uprm.edu/Data/prsn/RSS/catalogue/eqs1week.xml"
	if resp, err := client.Get(prsnURL); err == nil {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		// Decode ISO-8859-1 to UTF-8 properly, then update the XML declaration
		decoder := charmap.ISO8859_1.NewDecoder()
		utf8Bytes, _, _ := transform.Bytes(decoder, bodyBytes)
		bodyStr := strings.Replace(string(utf8Bytes), "ISO-8859-1", "UTF-8", 1)

		var prsnData struct {
			Channel struct {
				Items []struct {
					Title    string   `xml:"title"`
					Link     string   `xml:"link"`
					Lat      float64  `xml:"http://www.w3.org/2003/01/geo/wgs84_pos# lat"`
					Long     float64  `xml:"http://www.w3.org/2003/01/geo/wgs84_pos# long"`
					Subjects []string `xml:"http://purl.org/dc/elements/1.1/ subject"`
					Guid     string   `xml:"guid"`
					PubDate  string   `xml:"pubDate"`
				} `xml:"item"`
			} `xml:"channel"`
		}

		if err := xml.Unmarshal([]byte(bodyStr), &prsnData); err == nil {
			for _, item := range prsnData.Channel.Items {
				parts := strings.SplitN(item.Title, ",", 2)
				magStr := "0.0"
				loc := item.Title
				if len(parts) == 2 {
					magStr = strings.TrimSpace(strings.Replace(parts[0], "M", "", 1))
					loc = strings.TrimSpace(parts[1])
				}
				mag, _ := strconv.ParseFloat(magStr, 64)
				if mag <= 0 {
					log.Printf("PRSN: evento con magnitud 0.0 en %s (posiblemente preliminar)", loc)
				}

				// Find depth subject (the one ending in km)
				depth := "N/A"
				for _, s := range item.Subjects {
					if strings.Contains(s, "km") {
						depth = strings.TrimSpace(s)
						break
					}
				}

				t, err := time.Parse(time.RFC1123, strings.TrimSpace(item.PubDate))
				if err != nil {
					t = time.Now()
				}

				// Use guid + timestamp to avoid duplicate IDs from source feed
				id := fmt.Sprintf("prsn-%d-%s", t.UnixMilli(), strings.TrimSpace(item.Guid))

				allScraped = append(allScraped, Earthquake{
					ID:          id,
					Source:      "PRSN",
					Magnitude:   mag,
					Depth:       depth,
					Location:    loc,
					Time:        t.UnixMilli(),
					Coordinates: []float64{item.Lat, item.Long},
					URL:         strings.TrimSpace(item.Link),
				})
			}
			log.Printf("PRSN events parsed: %d", len(prsnData.Channel.Items))
		} else {
			log.Printf("Error decoding PRSN: %v", err)
		}
	}

	// 6. Fetch UWI SRC (Eastern Caribbean)
	uwiURL := "https://map.uwiseismic.com/data.php?format=leaflet&data=EQ&timeframe=3&bbox=-75,9,-55,16&nocache=" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	if req, err := http.NewRequest("GET", uwiURL, nil); err == nil {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("Referer", "https://map.uwiseismic.com/")
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyStr := strings.TrimSpace(string(bodyBytes))

			// UWI returns [] when empty, sometimes scalar on error
			if !strings.HasPrefix(bodyStr, "[") {
				if bodyStr != "" && bodyStr != "0" && bodyStr != "[]" {
					preview := bodyStr
					if len(preview) > 100 {
						preview = preview[:100]
					}
					log.Printf("UWI: respuesta inesperada: %s", preview)
				}
				log.Printf("UWI events parsed: 0")
			} else {
				var uwiData []struct {
					Name         string  `json:"name"`
					Lat          float64 `json:"lat"`
					Lon          float64 `json:"lon"`
					Magnitude    string  `json:"magnitude"`
					Depth        float64 `json:"depth"`
					Location     string  `json:"location"`
					NearbyCities string  `json:"nearby_cities"`
					TimestampUTC string  `json:"timestamp_utc"`
					DataType     string  `json:"data_type"`
				}
				if err := json.Unmarshal(bodyBytes, &uwiData); err == nil {
					parsed := 0
					for _, item := range uwiData {
						if item.DataType != "" && item.DataType != "eq" {
							continue
						}
						if item.Lat == 0 && item.Lon == 0 {
							continue
						}
						mag, _ := strconv.ParseFloat(strings.TrimSpace(item.Magnitude), 64)
						t, err := parseUWITimestamp(item.TimestampUTC)
						if err != nil {
							t = time.Now()
						}
						depth := "N/A"
						if item.Depth > 0 {
							depth = fmt.Sprintf("%.1f km", item.Depth)
						}
						loc := strings.TrimSpace(item.NearbyCities)
						if loc != "" {
							parts := strings.SplitN(loc, "~", 2)
							loc = strings.TrimSpace(parts[0])
						}
						if loc == "" {
							loc = strings.TrimSpace(item.Location)
						}
						if loc == "" {
							loc = "Caribe Oriental"
						}
						allScraped = append(allScraped, Earthquake{
							ID:          "uwi-" + strings.TrimSpace(item.Name),
							Source:      "UWI",
							Magnitude:   mag,
							Depth:       depth,
							Location:    loc,
							Time:        t.UnixMilli(),
							Coordinates: []float64{item.Lat, item.Lon},
							URL:         "https://uwiseismic.com/earthquakes/",
						})
						parsed++
					}
					log.Printf("UWI events parsed: %d", parsed)
				} else {
					log.Printf("Error decoding UWI: %v", err)
				}
			}
		}
	}

	// 7. Fetch IPGP (French Antilles)
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	ipgpURL := fmt.Sprintf("http://ws.ipgp.fr/fdsnws/event/1/query?starttime=%s&endtime=%s&minlatitude=14&maxlatitude=19&minlongitude=-63&maxlongitude=-59&format=text&minmagnitude=1.5&nodata=404", sevenDaysAgo, today)
	if resp, err := client.Get(ipgpURL); err == nil {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		text := string(bodyBytes)
		if strings.HasPrefix(text, "#EventID") {
			lines := strings.Split(text, "\n")
			header := strings.Split(lines[0], "|")
			idx := make(map[string]int)
			for i, h := range header {
				idx[strings.TrimSpace(h)] = i
			}
			parsed := 0
			for _, line := range lines[1:] {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.Split(line, "|")
				if len(parts) < len(header) {
					continue
				}
				eventID := strings.TrimSpace(parts[idx["#EventID"]])
				timeStr := strings.TrimSpace(parts[idx["Time"]])
				lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[idx["Latitude"]]), 64)
				lon, _ := strconv.ParseFloat(strings.TrimSpace(parts[idx["Longitude"]]), 64)
				depth, _ := strconv.ParseFloat(strings.TrimSpace(parts[idx["Depth/km"]]), 64)
				mag, _ := strconv.ParseFloat(strings.TrimSpace(parts[idx["Magnitude"]]), 64)
				loc := strings.TrimSpace(parts[idx["EventLocationName"]])

				if lat == 0 && lon == 0 {
					continue
				}

				t, err := parseIPGPTime(timeStr)
				if err != nil {
					t = time.Now()
				}

				if loc == "" {
					loc = "Antillas Francesas"
				}
				allScraped = append(allScraped, Earthquake{
					ID:          "ipgp-" + eventID,
					Source:      "IPGP",
					Magnitude:   math.Round(mag*10) / 10,
					Depth:       fmt.Sprintf("%.1f km", depth),
					Location:    loc,
					Time:        t.UnixMilli(),
					Coordinates: []float64{lat, lon},
					URL:         "https://ws.ipgp.fr/fdsnws/event/1/query?eventid=" + eventID,
				})
				parsed++
			}
			log.Printf("IPGP events parsed: %d", parsed)
		} else if resp.StatusCode != http.StatusNotFound {
			preview := text
			if len(preview) > 200 {
				preview = preview[:200]
			}
			log.Printf("IPGP unexpected response: %s", preview)
		}
	}

	// 8. Fetch SGC (Colombia)
	sgcURL := "https://apicatalogador.sgc.gov.co/api/events/search/"
	sevenDaysAgoSGC := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	todaySGC := time.Now().Format("2006-01-02")
	sgcPayload := fmt.Sprintf(`{"start_time":"%s","end_time":"%s","min_magnitude":1.0,"page":1,"page_size":100}`, sevenDaysAgoSGC, todaySGC)
	if req, err := http.NewRequest("POST", sgcURL, strings.NewReader(sgcPayload)); err == nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Origin", "https://www.sgc.gov.co")
		req.Header.Set("Referer", "https://www.sgc.gov.co/")
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			var sgcData struct {
				Count   int `json:"count"`
				Results struct {
					Success bool `json:"success"`
					Results []struct {
						ID        string  `json:"id"`
						Place     string  `json:"place"`
						UTCTime   string  `json:"utc_time"`
						Magnitude float64 `json:"magnitude"`
						MagType   string  `json:"mag_type"`
						Depth     float64 `json:"depth"`
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
						Agency    string  `json:"agency"`
					} `json:"results"`
				} `json:"results"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&sgcData); err == nil {
				parsed := 0
				for _, e := range sgcData.Results.Results {
					if e.Latitude == 0 && e.Longitude == 0 {
						continue
					}
					t, err := time.Parse("2006-01-02 15:04:05", e.UTCTime)
					if err != nil {
						t = time.Now()
					}
					loc := e.Place
					if loc == "" {
						loc = "Colombia"
					}
					depth := "N/A"
					if e.Depth > 0 {
						depth = fmt.Sprintf("%.1f km", e.Depth)
					}
					allScraped = append(allScraped, Earthquake{
						ID:          "sgc-" + e.ID,
						Source:      "SGC",
						Magnitude:   math.Round(e.Magnitude*10) / 10,
						Depth:       depth,
						Location:    loc,
						Time:        t.UnixMilli(),
						Coordinates: []float64{e.Latitude, e.Longitude},
						URL:         "https://www.sgc.gov.co/detallesismo/" + e.ID + "/resumen",
					})
					parsed++
				}
				log.Printf("SGC events parsed: %d", parsed)
			} else {
				log.Printf("Error decoding SGC: %v", err)
			}
		} else {
			log.Printf("Error fetching SGC: %v", err)
		}
	}

	// Load 7 days funvisis from DB
	funvHist := getFunvisisHistory7Days()
	allScraped = append(allScraped, funvHist...)

	var combined []Earthquake
	// Deduplicate all
	for _, incoming := range allScraped {
		duplicate := false
		for i, existing := range combined {
			timeDiff := math.Abs(float64(existing.Time - incoming.Time))
			if timeDiff <= 300000 { // 5 mins
				dist := getDistanceFromLatLonInKm(existing.Coordinates[0], existing.Coordinates[1], incoming.Coordinates[0], incoming.Coordinates[1])
				if dist < 50 { // 50 km
					if !strings.Contains(existing.Source, incoming.Source) {
						if incoming.Source == "FUNVISIS" || incoming.Source == "CSN" || incoming.Source == "PRSN" || incoming.Source == "SGC" {
							// Prefer local agencies for location, depth and magnitude
							combined[i].Location = incoming.Location
							combined[i].Magnitude = incoming.Magnitude
							combined[i].Depth = incoming.Depth
							combined[i].Source = incoming.Source + " + " + combined[i].Source
						} else {
							combined[i].Source = combined[i].Source + " + " + incoming.Source
						}
					}
					duplicate = true
					break
				}
			}
		}
		if !duplicate {
			combined = append(combined, incoming)
		}
	}

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Time > combined[j].Time
	})

	earthquakeMutex.Lock()
	// Detect if there's a new latest earthquake
	isNewQuake := false
	var newestQuake *Earthquake
	if len(cachedEarthquakes) > 0 && len(combined) > 0 {
		if combined[0].ID != cachedEarthquakes[0].ID && combined[0].Time > cachedEarthquakes[0].Time {
			isNewQuake = true
			newestQuake = &combined[0]
		}
	}
	cachedEarthquakes = combined
	earthquakeMutex.Unlock()

	if isNewQuake && newestQuake != nil {
		go func() {
			title := fmt.Sprintf("Nuevo Sismo: %.1f", newestQuake.Magnitude)
			body := fmt.Sprintf("%s\nProfundidad: %s\nFuente: %s", newestQuake.Location, newestQuake.Depth, newestQuake.Source)
			sendPushNotificationToAll(title, body)
		}()
	}

	broadcastSSE("earthquakes_update", map[string]interface{}{"earthquakes": combined})
}

// TELEGRAM SCRAPER
func fetchTelegramDataLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	fetchTelegramData()

	for range ticker.C {
		fetchTelegramData()
	}
}

func fetchTelegramData() {
	var allAlerts []TelegramAlert

	client := &http.Client{Timeout: 10 * time.Second}

	for _, ch := range telegramChannels {
		req, _ := http.NewRequest("GET", "https://t.me/s/"+ch.ID, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error scraping %s: %v", ch.ID, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		html := string(body)
		resp.Body.Close()

		segments := strings.Split(html, "data-post=\"")
		for i := 1; i < len(segments); i++ {
			seg := segments[i]

			idRe := regexp.MustCompile(`^([^"]+)"`)
			idMatch := idRe.FindStringSubmatch(seg)
			if idMatch == nil {
				continue
			}
			postId := idMatch[1]

			textRe := regexp.MustCompile(`(?s)<div class="tgme_widget_message_text[^>]*>(.*?)<\/div>`)
			textMatch := textRe.FindStringSubmatch(seg)
			if textMatch == nil {
				continue
			}

			rawText := textMatch[1]
			cleanText := strings.ReplaceAll(rawText, "<br/>", "\n")
			cleanText = strings.ReplaceAll(cleanText, "<br>", "\n")

			// Remove remaining HTML tags
			tagRe := regexp.MustCompile(`<[^>]+>`)
			cleanText = tagRe.ReplaceAllString(cleanText, "")
			cleanText = strings.TrimSpace(cleanText)

			lowerText := strings.ToLower(cleanText)

			alertType := ""
			if strings.Contains(lowerText, "desaparecid") || strings.Contains(lowerText, "fallecid") || strings.Contains(lowerText, "víctima") || strings.Contains(lowerText, "herid") || strings.Contains(lowerText, "accidente") || strings.Contains(lowerText, "vial") || strings.Contains(lowerText, "perdid") || strings.Contains(lowerText, "extraviad") || strings.Contains(lowerText, "se busca") {
				alertType = "MISSING"
			} else if strings.Contains(lowerText, "rescate") || strings.Contains(lowerText, "emergencia") || strings.Contains(lowerText, "apoyo") || strings.Contains(lowerText, "ayuda") || strings.Contains(lowerText, "insumo") || strings.Contains(lowerText, "incendio") || strings.Contains(lowerText, "inundación") || strings.Contains(lowerText, "inundacion") || strings.Contains(lowerText, "lluvias") || strings.Contains(lowerText, "tránsito") || strings.Contains(lowerText, "desbord") {
				alertType = "SUPPORT"
			} else if strings.Contains(lowerText, "sismo") || strings.Contains(lowerText, "terremoto") || strings.Contains(lowerText, "derrumbe") {
				alertType = "QUAKE"
			}

			if alertType != "" {
				img := ""
				imgRe := regexp.MustCompile(`background-image:url\('([^']+)'\)`)
				if m := imgRe.FindStringSubmatch(seg); m != nil {
					img = m[1]
				}

				dateStr := "Reciente"
				timeRe := regexp.MustCompile(`<time datetime="([^"]+)"`)
				if m := timeRe.FindStringSubmatch(seg); m != nil {
					dateStr = m[1]
				}

				shortText := cleanText
				if len(shortText) > 250 {
					shortText = shortText[:250] + "..."
				}

				allAlerts = append(allAlerts, TelegramAlert{
					Group:    ch.Name,
					Link:     "https://t.me/" + postId,
					Text:     shortText,
					FullText: cleanText,
					Type:     alertType,
					Image:    img,
					Date:     dateStr,
				})
			}
		}
	}

	sort.Slice(allAlerts, func(i, j int) bool {
		if allAlerts[i].Date == "Reciente" {
			return true
		}
		if allAlerts[j].Date == "Reciente" {
			return false
		}
		return allAlerts[i].Date > allAlerts[j].Date
	})

	if len(allAlerts) > 40 {
		allAlerts = allAlerts[:40]
	}

	alertMutex.Lock()
	cachedAlerts = allAlerts
	alertMutex.Unlock()

	broadcastSSE("telegram_update", map[string]interface{}{"alerts": allAlerts})
}

// VOLCANO ALERTS FETCHER
func fetchVolcanoAlertsLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run once immediately
	fetchVolcanoAlertsData()

	for range ticker.C {
		fetchVolcanoAlertsData()
	}
}

func fetchVolcanoAlertsData() {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://volcanoes.usgs.gov/vsc/api/volcanoApi/vhpstatus")
	if err != nil {
		log.Printf("Error fetching volcano alerts: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading volcano alerts response: %v", err)
		return
	}

	var rawVolcanoes []struct {
		Name        string  `json:"vName"`
		Lat         float64 `json:"lat"`
		Long        float64 `json:"long"`
		AlertLevel  string  `json:"alertLevel"`
		ColorCode   string  `json:"colorCode"`
		Synopsis    string  `json:"noticeSynopsis"`
		Threat      string  `json:"nvewsThreat"`
		Observatory string  `json:"obs"`
	}

	if err := json.Unmarshal(body, &rawVolcanoes); err != nil {
		log.Printf("Error decoding volcano alerts JSON: %v", err)
		return
	}

	var filtered []VolcanoAlert
	for _, v := range rawVolcanoes {
		level := strings.ToUpper(strings.TrimSpace(v.AlertLevel))
		if level == "ADVISORY" || level == "WATCH" || level == "WARNING" {
			filtered = append(filtered, VolcanoAlert{
				Name:        v.Name,
				Lat:         v.Lat,
				Lon:         v.Long,
				AlertLevel:  level,
				ColorCode:   v.ColorCode,
				Synopsis:    v.Synopsis,
				Threat:      v.Threat,
				Observatory: v.Observatory,
			})
		}
	}

	saveVolcanoAlerts(filtered)

	log.Printf("Alertas volcánicas actualizadas: %d volcanes con alerta", len(filtered))

	broadcastSSE("volcano_alerts_update", map[string]interface{}{"volcano_alerts": filtered})
}

func parseUWITimestamp(ts string) (time.Time, error) {
	ts = strings.TrimSpace(ts)
	layouts := []string{
		"2006-01-02 03:04:05 PM",
		"2006-01-02 03:04 PM",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, ts); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("no se pudo parsear timestamp UWI: %s", ts)
}

func parseIPGPTime(ts string) (time.Time, error) {
	ts = strings.TrimSpace(ts)
	layouts := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.999999999",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, ts); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("no se pudo parsear timestamp IPGP: %s", ts)
}
