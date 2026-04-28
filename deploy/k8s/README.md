# GoAdmin Kubernetes Manifests

This directory contains plain Kubernetes manifests for deploying GoAdmin.

It also includes Kustomize overlays for `dev`, `test`, and `prod` under `deploy/k8s/overlays/`.

## Usage

1. Create the namespace and secret:

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.example.yaml
```

2. Apply the rest of the manifests:

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
kubectl apply -f deploy/k8s/deployment-web.yaml
kubectl apply -f deploy/k8s/service-web.yaml
kubectl apply -f deploy/k8s/ingress.yaml
```

## Environment overlays

Use Kustomize overlays when you need environment-specific deployment settings:

```bash
kustomize build deploy/k8s/overlays/dev | kubectl apply -f -
kustomize build deploy/k8s/overlays/test | kubectl apply -f -
kustomize build deploy/k8s/overlays/prod | kubectl apply -f -
```

## Notes

- Replace the placeholder image reference before production use.
- Replace `GOADMIN_DATABASE_DSN` and `GOADMIN_AUTH_JWT_SECRET` with real secret values.
- The server container listens on port `8080` and serves `/api/v1/health` for health checks.
- The frontend container listens on port `80`, serves the SPA, and proxies `/api/` to the server service.
- `tenant.enabled` can be switched via `GOADMIN_TENANT_ENABLED`.
