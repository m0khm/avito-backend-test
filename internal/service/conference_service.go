package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ConferenceClient struct {
	baseURL string
	client  *http.Client
}

func NewConferenceClient(baseURL string, timeout time.Duration) *ConferenceClient {
	return &ConferenceClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

type conferenceRequest struct {
	BookingID string `json:"bookingId"`
	SlotID    string `json:"slotId"`
	UserID    string `json:"userId"`
}

type conferenceResponse struct {
	URL string `json:"url"`
}

func (c *ConferenceClient) CreateConferenceLink(ctx context.Context, bookingID, slotID, userID string) (*string, error) {
	payload, _ := json.Marshal(conferenceRequest{BookingID: bookingID, SlotID: slotID, UserID: userID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/conference-links", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build conference request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call conference service: %w", err)
	}
	defer func() {
	_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("conference service status: %d", resp.StatusCode)
	}
	var parsed conferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode conference service response: %w", err)
	}
	return &parsed.URL, nil
}
