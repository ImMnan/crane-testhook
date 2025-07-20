# crane-testhook

A Kubernetes cluster requirements checker for Blazemeter Private Locations. This tool verifies node resources, network connectivity, RBAC, and ingress configuration to ensure your cluster is ready to deploy Blazemeter workloads

## Features

- Checks node CPU, memory, and storage capacity
- Verifies network access to Blazemeter, Docker registry, and third-party endpoints
- Validates Kubernetes RBAC roles and bindings
- Confirms ingress or Istio gateway setup and TLS secret presence
- Designed to run as a Kubernetes Pod

## Prerequisites

- Go 1.23+
- Access to a Kubernetes cluster (with permissions to list nodes, roles, rolebindings, secrets, etc.)
- Docker (for building container images)

## Build

To build the binary for Linux/amd64:

```sh
go env -w GOOS=linux GOARCH=amd64
go build -o cranehook .
```

## Usage

### As a helm test hook

This image is integrated with helm test hook, so you can just run 
```sh
helm test <release-name>
```
This will automatically test the installation. 

### As a Kubernetes Pod

See [`kubernetes/cranehook.yaml`](kubernetes/cranehook.yaml) for an example manifest. Apply it with:

```sh
kubectl apply -f kubernetes/cranehook.yaml
```

The pod will run the checks and exit with code 0 if all requirements are met, or 1 if any check fails.

**NOTE**
- If you deploy crane-hook manually, with `kubectl apply` method, make sure the environment variables are set correctly as per the type of crane installation. See Environment Variables below.

## Environment Variables

- `WORKING_NAMESPACE`: Namespace to check for roles and resources
- `ROLE_NAME`: Name of the Role to check
- `ROLE_BINDING_NAME`: Name of the RoleBinding to check
- `SERVICE_ACCOUNT_NAME`: ServiceAccount to check in the RoleBinding
- `SV_ENABLE`: Set to `true` to enable ingress checks
- `KUBERNETES_WEB_EXPOSE_TYPE`: `INGRESS` or `ISTIO`
- `DOCKER_REGISTRY`: Docker registry URL to check
- `KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME`: TLS secret name for ingress/istio
- `KUBERNETES_ISTIO_GATEWAY_NAME`: (if using Istio) Gateway resource name
- `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`: (optional) Proxy settings

## Output

- `[INFO]` messages indicate successful checks
- `[error]` messages indicate failed checks
- Exit code 0: all checks passed
- Exit code 1: one or more checks failed (the logs would list the failures/errors)
