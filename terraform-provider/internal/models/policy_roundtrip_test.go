package models

import "testing"

func TestPolicyExpandFlattenRoundTrip(t *testing.T) {
	original := testPolicyModel()

	spec := ExpandPolicySpec(original)
	result := FlattenPolicySpec(spec)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Fatalf("round-trip name mismatch")
	}

	if result.Target.Service.ValueString() != original.Target.Service.ValueString() {
		t.Fatalf("round-trip service mismatch")
	}
}
