package main

import (
	"testing"
)

func TestGetDistanceFromLatLonInKm(t *testing.T) {
	// Distance between Caracas and Valencia (Venezuela) is roughly 125 km
	d := getDistanceFromLatLonInKm(10.48, -66.90, 10.17, -68.00)
	if d < 100 || d > 150 {
		t.Fatalf("expected distance between Caracas and Valencia ~125 km, got %.2f", d)
	}

	// Same point should be 0
	if getDistanceFromLatLonInKm(10.0, -66.0, 10.0, -66.0) != 0 {
		t.Fatal("expected 0 distance for same point")
	}
}

func TestParseUWITimestamp(t *testing.T) {
	ts, err := parseUWITimestamp("2024-05-01T12:34:56Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Year() != 2024 || ts.Month() != 5 || ts.Day() != 1 {
		t.Fatalf("unexpected parsed time: %v", ts)
	}
}

func TestParseIPGPTime(t *testing.T) {
	ts, err := parseIPGPTime("2024-05-01T12:34:56")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Year() != 2024 || ts.Month() != 5 || ts.Day() != 1 {
		t.Fatalf("unexpected parsed time: %v", ts)
	}
}
