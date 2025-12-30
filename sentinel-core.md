# Sentinel’s Core

Sentinel is a deployment guard that operates in three distinct phases. It is a controller, ideally deployed in an internal infrastructure cluster or a dedicated infra instance. Similar to Argo CD, Sentinel is configured via a manifest that defines which application to guard, which endpoints and SLOs apply, and which cluster the application runs in. Sentinel continuously observes application deployments and configuration changes within those clusters and evaluates whether change should be allowed to proceed.

The core philosophy behind Sentinel is derived directly from established SRE practices:

1. **Do not deploy change when you cannot observe its effects.**
   During outages, partial outages, or when telemetry is missing or unreliable, introducing change increases risk and complicates recovery. This phase does not judge service quality; it strictly determines whether the system is observable and stable enough for a deployment guard to make reliable decisions.

2. **Do not deploy when error budget consumption is unsafe.**
   A system may appear operational while still burning error budget at an unacceptable rate. Sentinel evaluates burn rate across multiple windows, remaining error budget, time to exhaustion, and the defined SLO to decide whether additional change would increase user harm. When reliability is already being overspent, deployment velocity must be reduced or halted.

3. **Rollback rapidly when new change causes user impact.**
   Sentinel treats new deployments as potentially unsafe until proven otherwise. After a canary rollout, Sentinel monitors error budget consumption attributable to the new version. If the change increases user-visible failures beyond acceptable thresholds, Sentinel triggers an immediate rollback. If no regression is detected, the new version is allowed to promote.

Together, these phases separate **change safety**, **organizational reliability intent**, and **change validation** into independent decisions. Sentinel does not attempt to judge correctness or performance in isolation; it exists solely to prevent deployments from increasing user harm when reliability signals indicate risk.

