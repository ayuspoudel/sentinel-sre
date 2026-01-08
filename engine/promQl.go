package engine

import (
	"fmt"
)

// buildBurnQuery constructs a PromQL expression that computes
// error budget burn rate for a given time window.
//
// errorsQuery: base error counter (no rate, no window)
// window:      PromQL window (e.g. "5m", "1h")
// allowedErr:  allowed error ratio derived from SLO (e.g. 0.001)
func buildBurnQuery(errorsQuery string, window string, allowedErr float64) string {
	return fmt.Sprintf(
		`((sum(rate(%s[%s])) OR vector(0)) / sum(rate(%s[%s]))) / %f`,
		errorsQuery,
		window,
		stripLabelFilters(errorsQuery),
		window,
		allowedErr,
	)
}
func stripLabelFilters(metric string) string {
	for i := 0; i < len(metric); i++ {
		if metric[i] == '{' {
			return metric[:i]
		}
	}
	return metric
}

func applyScope(query, label, value string) string {
	return query[:len(query)-1] + `,` + label + `="` + value + `"}` + query[len(query)-1:]
}
