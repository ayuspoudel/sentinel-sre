package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	tempDir := t.TempDir()
	// Creating nested Directory
	prodDir := filepath.Join(tempDir, "prod")
	if err := os.Mkdir(prodDir, 0755); err != nil {
		t.Fatalf("failed to create prod dir: %v", err)
	}
	manifestYaml := `
apiVersion: sentinel.sre/v1
kind: Guard
metadata:
  name: checkout-api
  owner: payments-sre
  environment: prod
target:
  cluster: prod
  namespace: checkout
  service: checkout
signals:
  traffic:
    source: prometheus
    query: sum(rate(http_requests_total[1m]))
    minRPS: 5
  errors:
    source: prometheus
    query: sum(rate(http_requests_total{status=~"5.."}[1m]))
  slo:
    objective: 99.9
    window: 30d
policy:
  budget:
    fastBurn:
      window: 5m
      threshold: 14
    slowBurn:
      window: 1h
      threshold: 2
    minRemaining: 10%`
	manifestPath := filepath.Join(prodDir, "checkout.yaml")
	err := os.WriteFile(manifestPath, []byte(manifestYaml), 0644)
	if err != nil {
		t.Fatalf("failed to write manifest : %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "README.txt"), []byte("ignore me"), 0644)
	if err != nil {
		t.Fatalf("failed to write non-yaml file: %v", err)
	}
	manifests, err := LoadManifests(tempDir)
	if err != nil {
		t.Fatalf("LoadManifests returned error: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}
	m := manifests[0]
	if m.Metadata.Name != "checkout-api" {
		t.Errorf("expected name 'checkout-api', got '%s'", m.Metadata.Name)
	}
	if m.Target.Cluster != "prod" {
		t.Errorf("expected cluster 'prod', got '%s'", m.Target.Cluster)
	}
	if m.Signals.Traffic.MinRPS != 5 {
		t.Errorf("expected minRPS 5, got %f", m.Signals.Traffic.MinRPS)
	}

}
