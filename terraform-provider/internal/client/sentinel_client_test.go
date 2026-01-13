package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
)

func TestSentinelClient_ApplyPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSentinelClient(server.URL, "")
	err := client.ApplyPolicy(context.Background(), "checkout", &spec.PolicySpec{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSentinelClient_GetPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewSentinelClient(server.URL, "")
	policy, err := client.GetPolicy(context.Background(), "missing")

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if policy != nil {
		t.Fatalf("expected nil policy")
	}
}
