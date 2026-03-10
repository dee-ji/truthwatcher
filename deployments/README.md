# deployments/

Deployment assets for local and cluster execution.

## Layout
- `docker/`: Dockerfiles for service binaries.
- `k8s/`: baseline Kubernetes manifests.
- `helm/truthwatcher/`: starter Helm chart.

Use architectural service names (`spanreed`, `squire`, `radiant`, `stormlight`, `seekers`, `highstorm`) consistently across deployment descriptors.

## TODO
- TODO(truthwatcher): align Helm and raw k8s manifests to expose the same configurable surfaces.
