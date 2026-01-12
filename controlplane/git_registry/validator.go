package registry

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

/*
@ayuspoudel
This functino takes a manifest loaded by loader.go's LoadManifest function
- If canary is enabled, canary.duration must be set
- If canary is enabled, canary.scope (label, canaryValue, baselineValue) must be fully defined
- fastBurn.threshold must be greater than slowBurn.threshold
*/
func ValidateManifest(m Manifest) error {
	err := validate.Struct(m)
	if err != nil {
		return err
	}
	return validateSemantics(m)
}

func validateSemantics(m Manifest) error {
	if m.Canary != nil && m.Canary.Enabled {
		if m.Canary.Duration == "" {
			return fmt.Errorf("canary.duration must be set when canary is enabled")
		}
		if m.Canary.Scope.Label == "" || m.Canary.Scope.CanaryValue == "" || m.Canary.Scope.BaselineValue == "" {
			return fmt.Errorf("canary.scope fields must be set when canary is enabled")
		}
	}
	if m.Policy.Budget.FastBurn.Threshold <= m.Policy.Budget.SlowBurn.Threshold {
		return fmt.Errorf("policy.budget.fastBurn.threshold must be greater than policy.budget.slowBurn.threshold")
	}
	return nil
}
