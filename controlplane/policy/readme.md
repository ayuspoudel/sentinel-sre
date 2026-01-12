# Policy Registry Service

The Policy Registry Service is responsible for managing and storing policies that govern the behavior of the control plane. It provides a centralized repository for policy definitions, allowing for easy retrieval, updates, and enforcement across different components of the system.

### Architecture

This follows a hexagonal architecture design pattern with implicit interface satisfation. 
The layers of design are

**Root**: Storage/ Persistence - `./store`

- It provides methods to persist, read and write data. Can be accessed using PolicyStore Interface.

**Level 1**: Specifications - `./spec`

- This defines what user input must have, how it will be read by policy registry, how other functions can access it. 
- It provides PolicySpec Interface. And also a validation() method.

**Core Level**: Service - `./service`
- This is the core business logic of policy registry. It provides methods to create, read, update, delete policies.
- It uses PolicyStore and PolicySpec interfaces to perform its operations.

**Adapter Layer**: Adapters - `./adapters`

- This layer contains implementations of the interfaces defined in the core and lower layers. Things like reading from other controller's db, other registry's db and quick validation. Any external agent this service talks to is defined here.

**API Layer**: API - `./api`

- This layer exposes the functionality of the policy registry service via RESTful APIs.
- Methods are:
    * /health
    * PUT /v1/policies/{name}
    * GET /v1/policies/{name}/status
    * DELETE /v1/policies/{name}


### API Documentation

#### Health

```
GET /health
```

Used for liveness checks.

Response

```
200 OK
ok
```



#### Apply Policy (Create / Update)

```
PUT /v1/policies/{name}
```

Creates or replaces a policy.

**Request Body**

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



#### Get Policy

```
GET /v1/policies/{name}
```

Returns the stored policy spec.

Response

```json
{ ...PolicySpec }
```



#### List Policies

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