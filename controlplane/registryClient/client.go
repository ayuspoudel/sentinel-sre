package registryClient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	http := &http.Client{Timeout: 5 * time.Second}
	return &Client{baseURL: baseURL, http: http}
}

func (c *Client) HealthCheck() error {
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %s", resp.Status)
	}
	return nil
}

func (c *Client) ListClusters(ctx context.Context) ([]*Cluster, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/clusters/list", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var clusters []*Cluster
	if err := json.NewDecoder(resp.Body).Decode(&clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}
