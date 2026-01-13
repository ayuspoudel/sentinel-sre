package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
)

/*
	Author: @ayuspoudel
	This struct & its methods are provider facing. Provider will provide
	attributes to these methods. It abstracts the core sentinel details
	like API endpoints, HTTP methods from provider.
	This is the heart of the provider. It provides following methods:
	- ApplyPolicy(ctx context.Context, name, policy *spec.PolicySpec) error
	- DeletePolicy(ctx context.Context, name string) error
	- GetPolicy(ctx context.Context, name string) (*spec.PolicySpec, error)

	Terraform resource will use this Sentinel Client to talk to sentinel
	without having to know the details of http about Sentinel Client
*/

type SentinelClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewSentinelClient(baseURL, token string) *SentinelClient {
	return &SentinelClient{baseURL: baseURL, token: token,
		http: &http.Client{Timeout: 10 * time.Second},
	}

}

func (s *SentinelClient) ApplyPolicy(ctx context.Context, name string, policy *spec.PolicySpec) error {
	body, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	apiEndpoint := fmt.Sprintf("%s/v1/policies/%s", s.baseURL, name)
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, apiEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	s.addHeaders(request)

	response, err := s.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusMultipleChoices {
		var body bytes.Buffer
		_, _ = body.ReadFrom(response.Body)
		return fmt.Errorf("failed to apply policy: %s %s", response.Status, body.String())
	}
	return nil
}

func (s *SentinelClient) GetPolicy(ctx context.Context, name string) (*spec.PolicySpec, error) {
	apiEndpoint := fmt.Sprintf("%s/v1/policies/%s", s.baseURL, name)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint, nil)
	if err != nil {
		return nil, err
	}
	s.addHeaders(request)

	response, err := s.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("failed to get policy: %s", response.Status)
	}
	var policy spec.PolicySpec
	err = json.NewDecoder(response.Body).Decode(&policy)
	if err != nil {
		return nil, err
	}
	return &policy, nil

}

func (s *SentinelClient) DeletePolicy(ctx context.Context, name string) error {
	apiEndpoint := fmt.Sprintf("%s/v1/policies/%s", s.baseURL, name)
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, apiEndpoint, nil)
	if err != nil {
		return err
	}
	s.addHeaders(request)
	response, err := s.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	if response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("failed to delete policy: %s", response.Status)
	}
	return nil

}

func (s *SentinelClient) addHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")

	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
}
