package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	SLOBurnRate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "sentinel",
			Help:      "Current burn rate of SLO",
			Subsystem: "slo",
			Name:      "burn_rate",
		})
	ErrorBudgetRemaining = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "sentinel",
		Help:      "Error budget remaining percentage",
		Subsystem: "slo",
		Name:      "error_budget_remaining",
	})
)

func Register() {
	prometheus.MustRegister(SLOBurnRate)
	prometheus.MustRegister(ErrorBudgetRemaining)
}
