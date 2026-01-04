# Executor Safety Model (Phase 3)

The executor safety model answers one question only:

> What is Sentinel allowed to touch, and what is it forbidden from touching?

If this is not nailed down, Sentinel becomes dangerous.



## Core safety principle (non-negotiable)

> Sentinel must only act on workloads that explicitly opt in.

No discovery-based mutation.
No “I found a Deployment so I’ll manage it.”
No heuristics.

Everything must be declared intent.



## Why opt-in is mandatory

Without opt-in:

* Sentinel could mutate prod workloads accidentally
* A bug could rollback unrelated services
* Cluster-wide RBAC becomes too permissive
* Users will not trust the tool

Real SRE tooling is paranoid by default.



## Safety boundary #1: Manifest is the contract

You already have this in your design, and it’s correct:

```yaml
apiVersion: sentinel.sre/v1
kind: Guard

metadata:
  name: checkout-api

target:
  cluster: minikube
  namespace: default
  service: checkout
```

This means:

* Sentinel is allowed to act only on:

  * cluster = minikube
  * namespace = default
  * service = checkout

Nothing else.

The executor must refuse to act outside this scope.



## Safety boundary #2: Label-based ownership (critical)

Sentinel must only mutate workloads that carry a sentinel ownership label.

Example (mandatory):

```yaml
metadata:
  labels:
    sentinel.guard: checkout-api
```

Rules:

* Executor ignores workloads without this label
* Label value must match `metadata.name` of the Guard
* One Guard → one workload ownership

This prevents:

* accidental takeover
* cross-guard interference
* “smart guessing” bugs



## Safety boundary #3: Namespace confinement

Executor must be namespace-scoped.

Even if RBAC allows cluster-wide access:

* Executor must refuse to touch:

  * other namespaces
  * cluster-scoped resources

This is enforced in code, not just RBAC.



## Safety boundary #4: Resource type allowlist

For v1, Sentinel executor is allowed to touch only:

* Deployment
* ReplicaSet (read-only)
* Service (selector only)

Explicitly forbidden:

* StatefulSet
* DaemonSet
* Job / CronJob
* Ingress
* ConfigMap
* Secret

If users want those later, it’s a v2 discussion.



## Safety boundary #5: One-way mutation rules

Sentinel must obey directional limits.

### Allowed:

* scale replicas down
* scale replicas up
* add labels
* remove canary deployment

### Forbidden:

* change container image
* change env vars
* change resource limits
* change ports
* change volumes

Sentinel does rollout control, not configuration management.



## Safety boundary #6: Canary isolation

Canary must be derived, never original.

Rules:

* Stable deployment is never mutated directly
* Canary is a copy with:

  * different name
  * different labels
  * reduced replicas

Rollback = delete canary
Promote = replace stable via controlled swap

This guarantees reversibility.



## Safety boundary #7: Idempotency

Executor operations must be safe to repeat.

Examples:

* “Create canary” when canary exists → no-op
* “Rollback” when no canary exists → no-op
* “Promote” when already promoted → no-op

This is critical because:

* engine ticks repeatedly
* actions may be re-emitted
* crashes can happen mid-operation



## Safety boundary #8: Action-driven execution only

Executor never polls metrics.
Executor never decides.

Executor only reacts to explicit actions:

```go
action.Rollback
action.Promote
action.Allow   // no-op
action.Block   // no-op
```

This keeps blast radius tiny.



## Safety boundary #9: Dry-run first (v1 requirement)

Executor must support dry-run mode:

```yaml
executor:
  dryRun: true
```

In dry-run:

* No mutations
* Full logging
* Same code paths

This allows:

* testing in prod safely
* confidence building
* CI validation



## Summary (memorize this)

Sentinel executor:

* acts only on declared targets
* requires explicit ownership labels
* is namespace-confined
* mutates only allowed resource types
* never changes config or images
* creates reversible canaries
* is idempotent
* executes actions, not logic
* supports dry-run

This is how you avoid writing a foot-gun.
