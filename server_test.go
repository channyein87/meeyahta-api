package main

import (
	"testing"
	"time"
)

func TestCleanStationName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with_suffix_and_platform", "Town Hall Station, Platform 3", "Town Hall"},
		{"without_comma", "Redfern Station", "Redfern"},
		{"no_suffix", "Museum", "Museum"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanStationName(tt.input); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestFormatTimeRFC3339(t *testing.T) {
	loc, _ := time.LoadLocation("Australia/Sydney")
	input := "2026-01-17T08:00:00Z"
	expected := time.Date(2026, 1, 17, 8, 0, 0, 0, time.UTC).In(loc).Format("03:04 PM")

	if got := formatTime(input, loc); got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestExtractTrips(t *testing.T) {
	loc, _ := time.LoadLocation("Australia/Sydney")
	resp := upstreamResponse{
		Journeys: []upstreamJourney{
			{
				Legs: []upstreamLeg{
					{
						Origin: waypoint{
							DisassembledName:       "Redfern Station, Platform 1",
							DepartureTimeEstimated: "2026-01-17T08:00:00Z",
						},
						Destination: waypoint{
							DisassembledName: "Central Station, Platform 20",
						},
					},
					{
						Origin: waypoint{
							DisassembledName: "Central Station, Platform 20",
						},
						Destination: waypoint{
							DisassembledName:     "Town Hall Station, Platform 5",
							ArrivalTimeEstimated: "2026-01-17T08:30:00Z",
						},
					},
				},
			},
		},
	}

	trips := extractTrips(resp, loc)
	if len(trips) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(trips))
	}

	expectedDeparture := time.Date(2026, 1, 17, 8, 0, 0, 0, time.UTC).In(loc).Format("03:04 PM")
	expectedArrival := time.Date(2026, 1, 17, 8, 30, 0, 0, time.UTC).In(loc).Format("03:04 PM")

	trip := trips[0]
	if trip.Origin != "Redfern" {
		t.Fatalf("expected origin %q, got %q", "Redfern", trip.Origin)
	}
	if trip.Destination != "Town Hall" {
		t.Fatalf("expected destination %q, got %q", "Town Hall", trip.Destination)
	}
	if trip.DepartureTime != expectedDeparture {
		t.Fatalf("expected departure %q, got %q", expectedDeparture, trip.DepartureTime)
	}
	if trip.ArrivalTime != expectedArrival {
		t.Fatalf("expected arrival %q, got %q", expectedArrival, trip.ArrivalTime)
	}
}
