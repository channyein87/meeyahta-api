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
							DisassembledName:       "Central Station, Platform 20",
							ArrivalTimePlanned:     "2026-01-17T08:10:00Z",
							ArrivalTimeBase:        "2026-01-17T08:11:00Z",
							ArrivalTimeEstimated:   "",
							DepartureTimeEstimated: "",
						},
					},
					{
						Origin: waypoint{
							DisassembledName:       "Central Station, Platform 20",
							DepartureTimePlanned:   "2026-01-17T08:15:00Z",
							DepartureTimeBase:      "2026-01-17T08:16:00Z",
							DepartureTimeEstimated: "",
						},
						Destination: waypoint{
							DisassembledName:     "Town Hall Station, Platform 5",
							ArrivalTimeEstimated: "2026-01-17T08:30:00Z",
							ArrivalTimePlanned:   "2026-01-17T08:31:00Z",
							ArrivalTimeBase:      "2026-01-17T08:32:00Z",
						},
					},
				},
			},
		},
	}

	trips := extractTrips(resp, loc)
	if len(trips) != 2 {
		t.Fatalf("expected 2 trips, got %d", len(trips))
	}

	expectedDeparture := time.Date(2026, 1, 17, 8, 0, 0, 0, time.UTC).In(loc).Format("03:04 PM")
	expectedArrival := time.Date(2026, 1, 17, 8, 10, 0, 0, time.UTC).In(loc).Format("03:04 PM")

	leg1 := trips[0]
	if leg1.Origin != "Redfern" || leg1.Destination != "Central" {
		t.Fatalf("unexpected leg1 origin/destination: %+v", leg1)
	}
	if leg1.Depart != expectedDeparture {
		t.Fatalf("expected leg1 depart %q, got %q", expectedDeparture, leg1.Depart)
	}
	if leg1.Arrive != expectedArrival {
		t.Fatalf("expected leg1 arrive %q, got %q", expectedArrival, leg1.Arrive)
	}

	expectedDeparture2 := time.Date(2026, 1, 17, 8, 15, 0, 0, time.UTC).In(loc).Format("03:04 PM")
	expectedArrival2 := time.Date(2026, 1, 17, 8, 30, 0, 0, time.UTC).In(loc).Format("03:04 PM")

	leg2 := trips[1]
	if leg2.Origin != "Central" || leg2.Destination != "Town Hall" {
		t.Fatalf("unexpected leg2 origin/destination: %+v", leg2)
	}
	if leg2.Depart != expectedDeparture2 {
		t.Fatalf("expected leg2 depart %q, got %q", expectedDeparture2, leg2.Depart)
	}
	if leg2.Arrive != expectedArrival2 {
		t.Fatalf("expected leg2 arrive %q, got %q", expectedArrival2, leg2.Arrive)
	}
}
