package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"
)

func initDB(db *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS funvisis_history (
		id TEXT PRIMARY KEY,
		magnitude REAL,
		depth TEXT,
		location TEXT,
		time INTEGER,
		lat REAL,
		lon REAL
	);

	CREATE TABLE IF NOT EXISTS push_subscriptions (
		endpoint TEXT PRIMARY KEY,
		keys_p256dh TEXT NOT NULL,
		keys_auth TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS volcano_alerts (
		name TEXT,
		lat REAL,
		lon REAL,
		alert_level TEXT,
		color_code TEXT,
		synopsis TEXT,
		threat TEXT,
		observatory TEXT,
		updated_at INTEGER
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	migrateHistoryIfNeeded(db)
}

func migrateHistoryIfNeeded(db *sql.DB) {
	// Check if history file exists
	fileData, err := os.ReadFile("funvisis_history.json")
	if err != nil {
		return // No history file, nothing to migrate
	}

	var history []Earthquake
	if err := json.Unmarshal(fileData, &history); err != nil {
		log.Printf("Error parsing funvisis_history.json: %v", err)
		return
	}

	// Begin transaction for fast insert
	tx, err := db.Begin()
	if err != nil {
		return
	}

	stmt, err := tx.Prepare(`INSERT INTO funvisis_history (id, magnitude, depth, location, time, lat, lon) VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`)
	if err != nil {
		return
	}
	defer stmt.Close()

	migrated := 0
	for _, eq := range history {
		var lat, lon float64
		if len(eq.Coordinates) == 2 {
			lat = eq.Coordinates[0]
			lon = eq.Coordinates[1]
		}
		res, err := stmt.Exec(eq.ID, eq.Magnitude, eq.Depth, eq.Location, eq.Time, lat, lon)
		if err == nil {
			affected, _ := res.RowsAffected()
			if affected > 0 {
				migrated++
			}
		}
	}
	tx.Commit()

	if migrated > 0 {
		log.Printf("Migrated %d historical records to SQLite.", migrated)
		os.Rename("funvisis_history.json", "funvisis_history.json.migrated")
	}
}

func saveFunvisisHistory(eq Earthquake) {
	insertSQL := `
	INSERT INTO funvisis_history (id, magnitude, depth, location, time, lat, lon)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO NOTHING;`

	_, err := db.Exec(insertSQL, eq.ID, eq.Magnitude, eq.Depth, eq.Location, eq.Time, eq.Coordinates[0], eq.Coordinates[1])
	if err != nil {
		log.Printf("Error saving to DB: %v", err)
	}
}

func getFunvisisHistory7Days() []Earthquake {
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour).UnixMilli()

	query := `SELECT id, magnitude, depth, location, time, lat, lon FROM funvisis_history WHERE time >= ? ORDER BY time DESC`
	rows, err := db.Query(query, sevenDaysAgo)
	if err != nil {
		log.Printf("Error reading 7 days history: %v", err)
		return nil
	}
	defer rows.Close()

	var history []Earthquake
	for rows.Next() {
		var eq Earthquake
		var lat, lon float64

		err = rows.Scan(&eq.ID, &eq.Magnitude, &eq.Depth, &eq.Location, &eq.Time, &lat, &lon)
		if err == nil {
			eq.Source = "FUNVISIS"
			eq.Coordinates = []float64{lat, lon}
			eq.URL = "http://www.funvisis.gob.ve/"
			eq.Felt = "Encuesta Disponible"
			eq.Alert = "Boletín Oficial"
			history = append(history, eq)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating 7 days history rows: %v", err)
	}
	return history
}

func getFunvisisHistoryByDate(dateStr string) []Earthquake {
	// Parse date target
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}

	startOfDay := targetDate.UnixMilli()
	endOfDay := targetDate.Add(24 * time.Hour).UnixMilli()

	return getFunvisisHistoryByRange(startOfDay, endOfDay)
}

func getFunvisisHistoryByDateRange(startDateStr, endDateStr string) []Earthquake {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil
	}
	endDate = endDate.Add(24 * time.Hour) // inclusive end day

	return getFunvisisHistoryByRange(startDate.UnixMilli(), endDate.UnixMilli())
}

func getFunvisisHistoryByRange(startMs, endMs int64) []Earthquake {
	query := `SELECT id, magnitude, depth, location, time, lat, lon FROM funvisis_history WHERE time >= ? AND time < ? ORDER BY time DESC`
	rows, err := db.Query(query, startMs, endMs)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var history []Earthquake
	for rows.Next() {
		var eq Earthquake
		var lat, lon float64

		err = rows.Scan(&eq.ID, &eq.Magnitude, &eq.Depth, &eq.Location, &eq.Time, &lat, &lon)
		if err == nil {
			eq.Source = "FUNVISIS"
			eq.Coordinates = []float64{lat, lon}
			eq.URL = "http://www.funvisis.gob.ve/"
			eq.Felt = "Encuesta Disponible"
			eq.Alert = "Boletín Oficial"
			history = append(history, eq)
		}
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return history
}

func saveVolcanoAlerts(alerts []VolcanoAlert) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting tx for volcano_alerts: %v", err)
		return
	}

	// Clear existing data
	if _, err := tx.Exec("DELETE FROM volcano_alerts"); err != nil {
		tx.Rollback()
		log.Printf("Error clearing volcano_alerts: %v", err)
		return
	}

	stmt, err := tx.Prepare(`INSERT INTO volcano_alerts (name, lat, lon, alert_level, color_code, synopsis, threat, observatory, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Printf("Error preparing volcano_alerts insert: %v", err)
		return
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for _, a := range alerts {
		if _, err := stmt.Exec(a.Name, a.Lat, a.Lon, a.AlertLevel, a.ColorCode, a.Synopsis, a.Threat, a.Observatory, now); err != nil {
			log.Printf("Error inserting volcano alert %s: %v", a.Name, err)
		}
	}

	tx.Commit()
}

func getVolcanoAlerts() []VolcanoAlert {
	rows, err := db.Query("SELECT name, lat, lon, alert_level, color_code, synopsis, threat, observatory FROM volcano_alerts")
	if err != nil {
		log.Printf("Error reading volcano_alerts: %v", err)
		return nil
	}
	defer rows.Close()

	var alerts []VolcanoAlert
	for rows.Next() {
		var a VolcanoAlert
		err = rows.Scan(&a.Name, &a.Lat, &a.Lon, &a.AlertLevel, &a.ColorCode, &a.Synopsis, &a.Threat, &a.Observatory)
		if err == nil {
			alerts = append(alerts, a)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating volcano_alerts: %v", err)
		return nil
	}
	return alerts
}
