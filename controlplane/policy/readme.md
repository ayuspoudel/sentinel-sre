

# Sentinel Policy Registry API

## Overview

The Policy Registry is an intent ingestion service for Sentinel SRE.

It stores what Sentinel should monitor and enforce, not how decisions are made.
It is designed to be consumed by Terraform, not humans.

This service:

* validates policy intent
* validates environment readiness (best-effort)
* persists policy spec and derived status
* exposes a read-only surface for decision systems

It does not evaluate policies.



## Core Concepts

### Policy Spec (Intent)

A policy describes:

* what workload Sentinel should watch
* which signals define health
* what SLO and error-budget rules apply

Policy spec is:

* declarative
* idempotent
* replace-on-write

### Policy Status (Observed)

Policy status describes:

* whether the environment is ready
* whether Sentinel can evaluate this policy
* why enforcement may be degraded

Status is:

* derived
* never user-authored
* non-blocking



## API Endpoints

### Health

```
GET /health
```

Used for liveness checks.

Response

```
200 OK
ok
```



### Apply Policy (Create / Update)

```
PUT /v1/policies/{name}
```

Creates or replaces a policy.

* Idempotent
* Full replace semantics
* Used by Terraform `apply`

Request Body

```json
{
  "metadata": {
    "owner": "payments-sre",
    "environment": "prod"
  },
  "target": {
    "cluster": "prod-us-east-1",
    "namespace": "checkout",
    "service": "checkout-api"
  },
  "signals": {
    "traffic": {
      "query": "sum(rate(http_requests_total[1m]))",
      "minRPS": 5
    },
    "errors": {
      "query": "sum(rate(http_requests_total{status=~\"5..\"}[1m]))"
    },
    "slo": {
      "objective": 99.9,
      "window": "30d"
    }
  },
  "policy": {
    "budget": {
      "fastBurn": {
        "window": "5m",
        "threshold": 14
      },
      "slowBurn": {
        "window": "1h",
        "threshold": 2
      }
    }
  }
}
```

Response

```json
{
  "status": "applied"
}
```

Notes

* `metadata.name` is derived from `{name}` path parameter
* Invalid intent is rejected immediately
* Environment issues do not block policy creation (captured in status)



### Get Policy

```
GET /v1/policies/{name}
```

Returns the stored policy spec.

Response

```json
{ ...PolicySpec }
```



### List Policies

```
GET /v1/policies
```

Returns all policies.

Response

```json
[
  { ...PolicySpec },
  { ...PolicySpec }
]
```



### Get Policy Status

```
GET /v1/policies/{name}/status
```

Returns derived policy status.

Response

```json
{
  "policy_name": "checkout-api",
  "cluster_exists": true,
  "cluster_reachable": true,
  "namespace_exists": true,
  "agent_installed": true,
  "agent_healthy": true,
  "prometheus_reachable": true,
  "queries_valid": true,
  "last_validated_at": "2026-01-11T21:40:00Z"
}
```

Notes

* Status explains *why* a policy may not be enforced
* Status does not affect policy existence



### Delete Policy

```
DELETE /v1/policies/{name}
```

Deletes policy and associated status.

Response

```
204 No Content
```

Used by Terraform `destroy`.



## Validation Model

Validation is split intentionally:

### Hard validation (blocks apply)

* policy spec structure
* SLO and budget semantics
* PromQL syntax
* cluster existence

### Soft validation (status only)

* namespace existence
* cluster reachability
* agent health
* Prometheus reachability

