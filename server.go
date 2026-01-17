package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type server struct {
	cfg      config
	client   *http.Client
	location *time.Location
}

type tripRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
}

type trip struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureTime string `json:"departureTime"`
	ArrivalTime   string `json:"arrivalTime"`
}

type tripResponse struct {
	Trips []trip `json:"trips"`
}

type upstreamResponse struct {
	Journeys []upstreamJourney `json:"journeys"`
}

type upstreamJourney struct {
	Legs []upstreamLeg `json:"legs"`
}

type upstreamLeg struct {
	Origin      waypoint `json:"origin"`
	Destination waypoint `json:"destination"`
}

type waypoint struct {
	DisassembledName       string `json:"disassembledName"`
	DepartureTimeEstimated string `json:"departureTimeEstimated"`
	ArrivalTimeEstimated   string `json:"arrivalTimeEstimated"`
}

func newServer(cfg config) *server {
	return &server{
		cfg:      cfg,
		client:   &http.Client{Timeout: 20 * time.Second},
		location: sydneyLocation(),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/trip", s.handleTrip)
	return mux
}

func (s *server) handleTrip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req tripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Origin = strings.TrimSpace(req.Origin)
	req.Destination = strings.TrimSpace(req.Destination)
	if req.Origin == "" || req.Destination == "" {
		http.Error(w, "origin and destination are required", http.StatusBadRequest)
		return
	}

	trips, err := s.fetchTrips(r.Context(), req.Origin, req.Destination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tripResponse{Trips: trips}); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

func (s *server) fetchTrips(ctx context.Context, origin, destination string) ([]trip, error) {
	resp, err := s.callTransportAPI(ctx, origin, destination)
	if err != nil {
		return nil, err
	}
	return extractTrips(resp, s.location), nil
}

func (s *server) callTransportAPI(ctx context.Context, origin, destination string) (upstreamResponse, error) {
	now := time.Now().In(s.location)
	q := url.Values{
		"outputFormat":      []string{"rapidJSON"},
		"coordOutputFormat": []string{"EPSG:4326"},
		"depArrMacro":       []string{"dep"},
		"itdDate":           []string{now.Format("20060102")},
		"itdTime":           []string{now.Format("1504")},
		"type_origin":       []string{"any"},
		"name_origin":       []string{origin},
		"type_destination":  []string{"any"},
		"name_destination":  []string{destination},
		"calcNumberOfTrips": []string{"2"},
		"excludedMeans":     []string{"checkbox"},
		"exclMOT_5":         []string{"1"},
		"TfNSWTR":           []string{"true"},
		"version":           []string{"10.2.1.42"},
		"itOptionsActive":   []string{"1"},
		"cycleSpeed":        []string{"16"},
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "api.transport.nsw.gov.au",
		Path:     "/v1/tp/trip",
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return upstreamResponse{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "apikey "+s.cfg.APIKey)

	res, err := s.client.Do(req)
	if err != nil {
		return upstreamResponse{}, fmt.Errorf("call transport nsw api: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return upstreamResponse{}, fmt.Errorf("transport nsw api returned status %d", res.StatusCode)
	}

	var body upstreamResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return upstreamResponse{}, fmt.Errorf("decode transport response: %w", err)
	}

	return body, nil
}

func extractTrips(resp upstreamResponse, loc *time.Location) []trip {
	trips := make([]trip, 0, len(resp.Journeys))
	for _, journey := range resp.Journeys {
		if len(journey.Legs) == 0 {
			continue
		}
		firstLeg := journey.Legs[0]
		lastLeg := journey.Legs[len(journey.Legs)-1]

		trips = append(trips, trip{
			Origin:        cleanStationName(firstLeg.Origin.DisassembledName),
			Destination:   cleanStationName(lastLeg.Destination.DisassembledName),
			DepartureTime: formatTime(firstLeg.Origin.DepartureTimeEstimated, loc),
			ArrivalTime:   formatTime(lastLeg.Destination.ArrivalTimeEstimated, loc),
		})
	}
	return trips
}

func cleanStationName(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.SplitN(name, ",", 2)
	base := strings.TrimSpace(parts[0])

	suffix := "Station"
	if len(base) >= len(suffix) && strings.EqualFold(base[len(base)-len(suffix):], suffix) {
		base = strings.TrimSpace(base[:len(base)-len(suffix)])
	}

	return base
}

func formatTime(value string, loc *time.Location) string {
	if value == "" {
		return ""
	}
	if loc == nil {
		loc = time.UTC
	}

	layouts := []string{
		time.RFC3339,
		"20060102T150405Z0700",
		"20060102T150405Z",
		"20060102T150405",
		"2006-01-02 15:04:05",
	}

	var parsed time.Time
	var parseErr error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			parsed = t
			parseErr = nil
			break
		} else {
			parseErr = err
		}
	}

	if parseErr != nil {
		return value
	}

	return parsed.In(loc).Format("03:04 PM")
}

func sydneyLocation() *time.Location {
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		log.Printf("warning: falling back to UTC timezone: %v", err)
		return time.UTC
	}
	return loc
}
