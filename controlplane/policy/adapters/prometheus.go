package adapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type PrometheusAdapter struct {
	baseURL string
	client  *http.Client
}

func NewPrometheusAdapter(baseURL string) *PrometheusAdapter {
	return &PrometheusAdapter{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *PrometheusAdapter) Query(ctx context.Context, query string) error {
	u := fmt.Sprintf("%s/api/v1/query?query=%s", p.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("prometheus query failed with status %d", resp.StatusCode)
	}
	return nil
}
