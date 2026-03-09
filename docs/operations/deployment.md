# Deployment

Truthwatcher deployments commonly include Radiant, Spanreed, Highstorm, Stormlight, Seekers, and Squire alongside PostgreSQL and Redis.

For local development, use `docker-compose.yml`.
For cluster deployments, start from `deployments/k8s` manifests or `deployments/helm/truthwatcher` and update images to your registry conventions.
