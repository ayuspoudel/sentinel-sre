package registry

type Manifest struct {
	APIVersion string `yaml:"apiVersion" validate:"required,eq=sentinel.sre/v1"`
	Kind       string `yaml:"kind" validate:"required,eq=Guard"`

	Metadata Metadata `yaml:"metadata" validate:"required"`
	Target   Target   `yaml:"target" validate:"required"`
	Signals  Signals  `yaml:"signals" validate:"required"`
	Policy   Policy   `yaml:"policy" validate:"required"`
	Canary   *Canary  `yaml:"canary"`
}

type Metadata struct {
	Name        string `yaml:"name" validate:"required"`
	Owner       string `yaml:"owner" validate:"required"`
	Environment string `yaml:"environment" validate:"required"`
}

type Target struct {
	Cluster   string `yaml:"cluster" validate:"required"`
	Namespace string `yaml:"namespace" validate:"required"`
	Service   string `yaml:"service" validate:"required"`
}

type Signals struct {
	Traffic TrafficSignal `yaml:"traffic" validate:"required"`
	Errors  ErrorSignal   `yaml:"errors" validate:"required"`
	SLO     SLO           `yaml:"slo" validate:"required"`
}

type TrafficSignal struct {
	Source string  `yaml:"source" validate:"required"`
	Query  string  `yaml:"query" validate:"required"`
	MinRPS float64 `yaml:"minRPS" validate:"gte=0"`
}

type ErrorSignal struct {
	Source string `yaml:"source" validate:"required"`
	Query  string `yaml:"query" validate:"required"`
}

type SLO struct {
	Objective float64 `yaml:"objective" validate:"gt=0,lt=100"`
	Window    string  `yaml:"window" validate:"required"`
}

type Policy struct {
	Budget BudgetPolicy `yaml:"budget" validate:"required"`
}

type BudgetPolicy struct {
	FastBurn     BurnWindow `yaml:"fastBurn" validate:"required"`
	SlowBurn     BurnWindow `yaml:"slowBurn" validate:"required"`
	MinRemaining string     `yaml:"minRemaining" validate:"required"`
}

type BurnWindow struct {
	Window    string  `yaml:"window" validate:"required"`
	Threshold float64 `yaml:"threshold" validate:"gt=0"`
}

type Canary struct {
	Enabled  bool        `yaml:"enabled"`
	Duration string      `yaml:"duration" validate:"required_if=Enabled true"`
	Scope    CanaryScope `yaml:"scope" validate:"required_if=Enabled true"`
}

type CanaryScope struct {
	Label         string `yaml:"label"`
	CanaryValue   string `yaml:"canaryValue"`
	BaselineValue string `yaml:"baselineValue"`
}
