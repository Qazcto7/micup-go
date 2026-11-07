package micup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	MicupListAPIBase = "https://api.micuplist.me/v1"
)

// MicupListClient provides helper methods for MicupList API integration.
type MicupListClient struct {
	apiToken   string
	httpClient *http.Client
}

// NewMicupListClient initializes a client with the MicupList API token.
func NewMicupListClient(apiToken string) *MicupListClient {
	return &MicupListClient{
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// StatsPayload payload for updating bot server count.
type StatsPayload struct {
	ServerCount int `json:"server_count"`
	ShardCount  int `json:"shard_count,omitempty"`
}

// PostStats posts bot server count to MicupList.
func (m *MicupListClient) PostStats(ctx context.Context, botID string, serverCount int) error {
	payload := StatsPayload{ServerCount: serverCount}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/bots/%s/stats", MicupListAPIBase, botID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", m.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("micuplist api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// VoteCheckResponse result of checking if user voted.
type VoteCheckResponse struct {
	Voted   bool      `json:"voted"`
	VotedAt time.Time `json:"voted_at,omitempty"`
}

// CheckUserVote checks if a specific user has voted for the bot.
func (m *MicupListClient) CheckUserVote(ctx context.Context, botID, userID string) (*VoteCheckResponse, error) {
	reqURL := fmt.Sprintf("%s/bots/%s/votes/%s", MicupListAPIBase, botID, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", m.apiToken)
	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("micuplist check vote failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var res VoteCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

// VoteWebhookPayload payload sent by MicupList vote webhook.
type VoteWebhookPayload struct {
	Type      string `json:"type"` // "vote" or "test"
	BotID     string `json:"bot_id"`
	UserID    string `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
	IsWeekend bool   `json:"is_weekend,omitempty"`
}

// NewVoteWebhookHandler creates an http.HandlerFunc to receive vote webhooks.
func NewVoteWebhookHandler(secret string, onVote func(payload *VoteWebhookPayload)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if secret != "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != secret {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		var payload VoteWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if onVote != nil {
			go onVote(&payload)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
