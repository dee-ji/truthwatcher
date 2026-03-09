# deployments/

Deployment assets for Truthwatcher.

- `docker/`: container build definitions for service binaries.
- `k8s/`: baseline Kubernetes manifests.
- `helm/truthwatcher/`: starter Helm chart.

These assets use architectural service names (`spanreed`, `squire`, `radiant`, `stormlight`) and should stay aligned with `cmd/` entrypoints.
