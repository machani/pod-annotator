# Pod Annotator

A small Kubebuilder/controller-runtime project that watches Kubernetes Pods and adds an annotation when a Pod matches a configured namespace and label.

This project intentionally **does not define a CRD**. It reconciles the built-in Kubernetes `Pod` resource and is intended as a focused example for learning controller-runtime fundamentals.

## Behavior

By default a Pod is selected when:

- namespace: `maas-t-tenant5`
- label: `prometheus=maas-t-tenant5-prometheus`

The controller then ensures this annotation exists:

```yaml
telemetry-compass.com/reconciled: "true"
```

Existing labels and annotations are preserved. Reconciliation is idempotent: a Pod that already has the desired annotation is not modified again.

```text
Pod event
   |
   v
namespace matches? -- no --> ignore
   |
  yes
   v
label matches? ----- no --> ignore
   |
  yes
   v
annotation already correct? -- yes --> no-op
   |
  no
   v
patch Pod annotation
```

## Prerequisites

- Go version specified in `go.mod`
- Docker or another compatible container builder
- `kubectl`
- access to a Kubernetes cluster
- GNU Make

Kubebuilder helper binaries used by the Makefile are downloaded into `bin/` as needed.

## Run locally

Use your current kubeconfig and run the manager outside the cluster:

```bash
make run
```

The matching behavior can be changed without rebuilding:

```bash
go run ./cmd/main.go \
  --target-namespace=maas-t-tenant5 \
  --target-label-key=prometheus \
  --target-label-value=maas-t-tenant5-prometheus \
  --annotation-key=telemetry-compass.com/reconciled \
  --annotation-value=true
```

## Test

Run formatting, vetting and envtest-based controller tests:

```bash
make test
```

The controller tests cover matching Pods, wrong/missing labels, wrong namespaces, preservation of unrelated annotations, idempotency and deleted Pods.

Run the Kind-based end-to-end suite with:

```bash
make test-e2e
```

## Try it manually

Create the target namespace if needed:

```bash
kubectl create namespace maas-t-tenant5
```

Create a matching Pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: annotated-demo
  namespace: maas-t-tenant5
  labels:
    prometheus: maas-t-tenant5-prometheus
spec:
  containers:
    - name: pause
      image: registry.k8s.io/pause:3.10
```

Then verify the annotation:

```bash
kubectl get pod annotated-demo -n maas-t-tenant5 \
  -o jsonpath='{.metadata.annotations.telemetry-compass\.com/reconciled}{"\n"}'
```

Expected output:

```text
true
```

## Build and deploy

Build and push an image to a registry accessible by your cluster:

```bash
make docker-build docker-push IMG=ghcr.io/machani/pod-annotator:v0.1.0
```

Deploy it:

```bash
make deploy IMG=ghcr.io/machani/pod-annotator:v0.1.0
```

Do not deploy the default `controller:latest` image unless that image is deliberately available to the cluster. Otherwise Kubernetes may try to pull `docker.io/library/controller:latest`.

Check the manager:

```bash
kubectl get pods -n pod-annotator-system
kubectl logs -n pod-annotator-system deployment/pod-annotator-controller-manager
```

Remove the deployment:

```bash
make undeploy
```

## Runtime configuration

The manager supports these controller-specific flags:

| Flag | Default |
| --- | --- |
| `--target-namespace` | `maas-t-tenant5` |
| `--target-label-key` | `prometheus` |
| `--target-label-value` | `maas-t-tenant5-prometheus` |
| `--annotation-key` | `telemetry-compass.com/reconciled` |
| `--annotation-value` | `true` |

The deployment defaults are in `config/manager/manager.yaml`.

## RBAC

The controller only needs to read Pods and patch/update their metadata. RBAC is generated from the marker in `internal/controller/pod_controller.go`:

```go
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;patch;update
```

Regenerate manifests after changing RBAC markers:

```bash
make manifests
```

## CI and image publishing

GitHub Actions run linting, envtest tests and E2E tests. `build-image.yml` publishes tagged releases to GitHub Container Registry (GHCR), using the Git tag as the image tag.

For example, pushing Git tag `v0.1.0` publishes:

```text
ghcr.io/machani/pod-annotator:v0.1.0
```

## Project structure

```text
cmd/main.go                         manager startup and command-line flags
internal/controller/               reconciliation logic and envtest tests
config/manager/                    Kubernetes Deployment
config/rbac/                       generated RBAC manifests
test/e2e/                          Kind-based end-to-end tests
.github/workflows/                 CI and image publishing
```
