package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type conferenceRequest struct {
	BookingID string `json:"bookingId"`
}

type conferenceResponse struct {
	URL string `json:"url"`
}

func main() {
	port := env("CONFERENCE_MOCK_PORT", "8090")
	failRate := envFloat("CONFERENCE_MOCK_FAIL_RATE", 0)
	mux := http.NewServeMux()
	mux.HandleFunc("/conference-links", func(w http.ResponseWriter, r *http.Request) {
		if rand.Float64() < failRate {
			http.Error(w, `{"error":"mock failure"}`, http.StatusServiceUnavailable)
			return
		}
		var req conferenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(conferenceResponse{URL: fmt.Sprintf("https://meet.mock.local/%s", req.BookingID)})
	})
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 3 * time.Second}
	log.Printf("conference mock listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return fallback
}
