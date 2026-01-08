package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Action struct {
	Type   string `json:"Type"`
	Reason string `json:"Reason"`
}

type SentinelClient struct {
	baseURL string
	client  *http.Client
}

func NewSentinelClient() (*SentinelClient, error) {
	url := os.Getenv("SENTINEL_SRE_URL")
	if url == "" {
		return nil, fmt.Errorf("SENTINEL_SRE_URL is not set")
	}

	return &SentinelClient{
		baseURL: url,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}, nil
}

func (s *SentinelClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		s.baseURL+"/health",
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sentinel unhealthy: %s", resp.Status)
	}

	return nil
}

func (s *SentinelClient) CheckWithSentinel(ctx context.Context, guard string) (bool, string) {
	if guard == "" {
		return false, "no guard specified"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/actions/"+guard, nil)
	if err != nil {
		return false, "failed to create request: " + err.Error()
	}

	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return false, "sentinel unreachable"
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, "sentinel returned non-200 status: " + resp.Status
	}
	var action Action
	err = json.NewDecoder(resp.Body).Decode(&action)
	if err != nil {
		return false, "invalid response from sentinel " + err.Error()
	}
	switch action.Type {
	case "allow":
		return true, action.Reason
	case "deny":
		return false, action.Reason
	default:
		return false, "unknown response from sentinel"
	}
}
