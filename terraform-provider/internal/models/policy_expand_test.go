package models

import "testing"

func TestExpandPolicySpec(t *testing.T) {
	tests := []struct {
		name  string
		model PolicyModel
	}{
		{
			name:  "valid full policy",
			model: testPolicyModel(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := ExpandPolicySpec(tt.model)

			if spec.Metadata.Name != "checkout" {
				t.Fatalf("expected name checkout, got %s", spec.Metadata.Name)
			}

			if spec.Target.Cluster != "sreCluster" {
				t.Fatalf("cluster mismatch")
			}

			if spec.Signals.Traffic.MinRPS != 10 {
				t.Fatalf("minRPS mismatch")
			}

			if spec.Policy.Budget.FastBurn.Threshold != 2 {
				t.Fatalf("fast burn threshold mismatch")
			}
		})
	}
}
