package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConferenceClient_CreateConferenceLink(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/conference-links" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"https://meet.test/link-1"}`))
	}))
	defer srv.Close()

	client := NewConferenceClient(srv.URL, time.Second)

	link, err := client.CreateConferenceLink(context.Background(), "booking-1", "slot-1", "user-1")
	if err != nil {
		t.Fatalf("CreateConferenceLink() error = %v", err)
	}
	if link == nil || *link != "https://meet.test/link-1" {
		t.Fatalf("unexpected link: %+v", link)
	}
}

func TestConferenceClient_CreateConferenceLink_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer srv.Close()

	client := NewConferenceClient(srv.URL, time.Second)

	if _, err := client.CreateConferenceLink(context.Background(), "booking-1", "slot-1", "user-1"); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}