package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMain(m *testing.M) {
	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	initDB(db)
	m.Run()
	db.Close()
}

func TestVapidKeyHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/vapidPublicKey", nil)
	rec := httptest.NewRecorder()

	vapidKeyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["publicKey"] != vapidPublicKey {
		t.Fatalf("expected public key %q, got %q", vapidPublicKey, body["publicKey"])
	}
}

func TestVolcanoAlertsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/volcanoAlerts", nil)
	rec := httptest.NewRecorder()

	volcanoAlertsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", contentType)
	}

	var body []VolcanoAlert
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body["status"])
	}
}

func TestSubscribeHandler(t *testing.T) {
	payload := map[string]interface{}{
		"endpoint": "https://fcm.googleapis.com/fcm/send/test",
		"keys": map[string]string{
			"p256dh": "test-p256dh",
			"auth":   "test-auth",
		},
	}
	data, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", bytes.NewReader(data))
	rec := httptest.NewRecorder()

	subscribeHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}
