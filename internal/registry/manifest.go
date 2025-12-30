package registry

type Manifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`

	Metadata Metadata `yaml:"metadata"`
	Target   Target   `yaml:"target"`
	Signals  Signals  `yaml:"signals"`
	Policy   Policy   `yaml:"policy"`
	Canary   *Canary  `yaml:"canary,omitempty"`
}

type Metadata struct {
	Name        string `yaml:"name"`
	Owner       string `yaml:"owner"`
	Environment string `yaml:"environment"`
}

type Target struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
	Service   string `yaml:"service"`
}

type Signals struct {
	Traffic TrafficSignal `yaml:"traffic"`
	Errors  ErrorSignal   `yaml:"errors"`
	SLO     SLO           `yaml:"slo"`
}

type TrafficSignal struct {
	Source string  `yaml:"source"`
	Query  string  `yaml:"query"`
	MinRPS float64 `yaml:"minRPS"`
}

type ErrorSignal struct {
	Source string `yaml:"source"`
	Query  string `yaml:"query"`
}

type SLO struct {
	Objective float64 `yaml:"objective"`
	Window    string  `yaml:"window"`
}

type Policy struct {
	Budget BudgetPolicy `yaml:"budget"`
}

type BudgetPolicy struct {
	FastBurn     BurnWindow `yaml:"fastBurn"`
	SlowBurn     BurnWindow `yaml:"slowBurn"`
	MinRemaining string     `yaml:"minRemaining"`
}

type BurnWindow struct {
	Window    string  `yaml:"window"`
	Threshold float64 `yaml:"threshold"`
}

type Canary struct {
	Enabled  bool        `yaml:"enabled"`
	Duration string      `yaml:"duration"`
	Scope    CanaryScope `yaml:"scope"`
}

type CanaryScope struct {
	Label         string `yaml:"label"`
	CanaryValue   string `yaml:"canaryValue"`
	BaselineValue string `yaml:"baselineValue"`
}
