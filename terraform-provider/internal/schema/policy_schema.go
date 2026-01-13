package schema

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

/*

Author: @ayushpoudel
This file defines the Terraform schema for the Sentinel policy resource.
The schema describes the structure of the resource block as it appears in Terraform
configuration and validates user input. Nested blocks and attributes defined here
map directly to the policy model used by the resource implementation.
*/

type PolicySchemaDef struct {
}

func MetadataSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"owner":       schema.StringAttribute{Required: true},
			"environment": schema.StringAttribute{Required: true},
		},
	}
}

func TargetSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"cluster":   schema.StringAttribute{Required: true},
			"namespace": schema.StringAttribute{Required: true},
			"service":   schema.StringAttribute{Required: true},
		},
	}
}

func SignalsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"traffic": trafficSignalSchema(),
			"errors":  errorSignalSchema(),
			"slo":     sloSchema(),
		},
	}
}

func PolicySchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"budget": budgetSchema(),
		},
	}
}

func trafficSignalSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"query":   schema.StringAttribute{Required: true},
			"min_rps": schema.Float64Attribute{Optional: true},
		},
	}
}

func errorSignalSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{Required: true},
		},
	}
}

func sloSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"objective": schema.Float64Attribute{Required: true},
			"window":    schema.StringAttribute{Required: true},
		},
	}
}

func budgetSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"fast_burn": burnWindowSchema(),
			"slow_burn": burnWindowSchema(),
		},
	}
}

func burnWindowSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"window":    schema.StringAttribute{Required: true},
			"threshold": schema.Float64Attribute{Required: true},
		},
	}
}

/*
Example usage in Terraform:

resource "sentinel_policy" "example" {
  metadata {
    owner       = "team-a"
    environment = "prod"
  }

  target {
    cluster   = "prod-cluster"
    namespace = "default"
    service   = "checkout"
  }

  signals {
    traffic {
      query   = "sum(rate(http_requests_total[1m]))"
      min_rps = 10
    }

    errors {
      query = "sum(rate(http_requests_errors_total[1m]))"
    }

    slo {
      objective = 99.9
      window    = "30d"
    }
  }

  policy {
    budget {
      fast_burn {
        window    = "5m"
        threshold = 2.0
      }

      slow_burn {
        window    = "1h"
        threshold = 1.0
      }
    }
  }
}
*/
