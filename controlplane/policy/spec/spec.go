package spec

/*
	Author: @ayuspoudel
	This file is the base of policy registry. It defines what we accept as policy.
	Policy Spec is the root object the whole thing looks like
		PolicySpec
			Metadata (Defined in Metadata struct, received as "metadata")
				|- Name          string
				|- Owner         string
				|- Environment   string

			Target (Defined in Target struct, received as "target")
				|- Cluster       string
				|- Namespace     string
				|- Service       string

			Signals (Defined in Signals struct, received as "signals")
				|- Traffic (TrafficSignal)
					|- Query     string
					|- MinRPS    float64
				|- Errors (ErrorSignal)
					|- Query     string
				|- SLO (SLO)
					|- Objective float64
					|- Window    string

			Policy (Defined in Policy struct, received as "policy")
				|- Budget (BudgetPolicy)
					|- FastBurn (BurnWindow)
						|- Window    string
						|- Threshold float64
					|- SlowBurn (BurnWindow)
						|- Window    string
						|- Threshold float64
*/

type PolicySpec struct {
	Metadata Metadata `json:"metadata" yaml:"metadata"`
	Target   Target   `json:"target" yaml:"target"`
	Signals  Signals  `json:"signals" yaml:"signals"`
	Policy   Policy   `json:"policy" yaml:"policy"`
}

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Owner       string `json:"owner" yaml:"owner"`
	Environment string `json:"environment" yaml:"environment"`
}

type Target struct {
	Cluster   string `json:"cluster" yaml:"cluster"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
}

type Signals struct {
	Traffic TrafficSignal `json:"traffic" yaml:"traffic"`
	Errors  ErrorSignal   `json:"errors" yaml:"errors"`
	SLO     SLO           `json:"slo" yaml:"slo"`
}

type TrafficSignal struct {
	Query  string  `json:"query" yaml:"query"`
	MinRPS float64 `json:"minRPS" yaml:"minRPS"`
}

type ErrorSignal struct {
	Query string `json:"query" yaml:"query"`
}

type SLO struct {
	Objective float64 `json:"objective" yaml:"objective"`
	Window    string  `json:"window" yaml:"window"`
}

type Policy struct {
	Budget BudgetPolicy `json:"budget" yaml:"budget"`
}

type BudgetPolicy struct {
	FastBurn BurnWindow `json:"fastBurn" yaml:"fastBurn"`
	SlowBurn BurnWindow `json:"slowBurn" yaml:"slowBurn"`
}

type BurnWindow struct {
	Window    string  `json:"window" yaml:"window"`
	Threshold float64 `json:"threshold" yaml:"threshold"`
}
