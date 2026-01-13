package models

import (
	"testing"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
)

func TestFlattenPolicySpec(t *testing.T) {
	tests := []struct {
		name string
		spec *spec.PolicySpec
	}{
		{
			name: "valid sentinel policy",
			spec: ExpandPolicySpec(testPolicyModel()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := FlattenPolicySpec(tt.spec)

			if model.Name.ValueString() != "checkout" {
				t.Fatalf("expected checkout")
			}

			if model.Policy.Budget.FastBurn.Threshold.ValueFloat64() != 2 {
				t.Fatalf("threshold mismatch")
			}
		})
	}
}
