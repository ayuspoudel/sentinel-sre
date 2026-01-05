# Sentinel Agent Helm Chart

This agent is a k8s admissions controller that enforces SRE safety gates during deployments. It blocks or allows deployments by consulting a central Sentinel SRE control plane, based on observability and error-budget health.

This chart installs sentinel agent in a k8s cluster.

### What this agent does?

- Runs inside your Kubernetes cluster
- Implements a Validating Admission Webhook
- Intercepts CREATE / UPDATE of Deployment resources
- Calls Sentinel SRE to decide whether a deployment is safe
- Blocks deployments when SRE safety rules are violated

## Installation

```bash
helm repo add sentinel https://example.com/sentinel/charts
helm repo update
```

**Install sentinel agent**

```bash
helm install sentinel-agent sentinel/sentinel-agent \
  --namespace sentinel \
  --create-namespace \
  --set sentinelSRE.url=https://sentinel-sre.company.internal \
  --set tls.caBundle=<BASE64_CA_CERT>
```

## Prerequisites

- Kubernetes v1.22+
- A running sentinel control plane
- TLS Certificates for admission webhooks
- Helm v3

### Values

| Key                      | Description                         | Default              |
| ------------------------ | ----------------------------------- | -------------------- |
| `namespace.name`         | Namespace to install Sentinel Agent | `sentinel`           |
| `namespace.create`       | Create namespace if missing         | `true`               |
| `image.repository`       | Agent image repository              | `sentinel-agent`     |
| `image.tag`              | Agent image tag                     | `latest`             |
| `replicaCount`           | Number of agent replicas            | `1`                  |
| `sentinelSRE.url`        | Sentinel SRE API endpoint           | required             |
| `tls.secretName`         | TLS secret name                     | `sentinel-agent-tls` |
| `tls.caBundle`           | Base64 encoded CA cert              | required             |
| `webhook.failurePolicy`  | Webhook failure policy              | `Fail`               |
| `webhook.timeoutSeconds` | Webhook timeout                     | `5`                  |
