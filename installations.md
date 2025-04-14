# Setup and Installation

### CLI

Install CLI application from Homebrew:

```bash
brew tap CS6510-SEA-SP25/pipeci
brew install pipeci
```

To upgrade version:

```bash
brew upgrade pipeci
```

To verify the CLI is working:

```shell
pipeci help
```

### Backend

For this version, the only way to run the project on local is using minikube. We recommend you to increase the resouce capacity of Docker Desktop to at least 6 CPUs and 8GB Memory to avoid crashing or resource limit exceeded while running.

Please follow the instructions at root directory.

To setup for the minikube cluster.

### 1. Start the minikube cluster

```bash
minikube start --cpus=6 --memory=8g
```

### 2. Secrets

We need 3 Kubernetes secrets, Github, MySQL, and Redis:

`k8s/github-secret.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-secret
type: Opaque
data:
  token: [value from:"echo -n $GITHUB_TOKEN | base64"]
```

`k8s/mysql-secret.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysql-secret
type: Opaque
data:
  DB_PASSWORD: [value from:"echo -n ... | base64"]
```

`k8s/redis-secret.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-secret
type: Opaque
data:
  REDIS_PASSWORD: [value from:"echo -n ... | base64"]
```

### 3. CA Certificate

We are using [Aiven](https://aiven.io/mysql) as the managed service for MySQL. As the hosted database is only accessible with SSL, we need a CA certificate file. Create a `ca.pem` file and place at the root directory.

```text
-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
```

### 4. Deploy to minikube

```bash
./deploy.sh
```

To check for status, please run

```bash
kubectl get pods
```

To expose the LoadBalancer services to localhost, run this on a different terminal window or in background.

```bash
minikube tunnel
```

NOTE: Please wait until all pods runs successfully before moving on to the next step.

### 5. Deploy operator

Deploy the controller to the cluster. This has to be done with _multiple terminal windows_.

Change directory into the operator codebase

```bash
cd operator/hpa
```

Start the operator controller manager

<details>
<summary>Details for the make command</summary>

**Generate the CustomResourceDefinitions(CRD) manifests**

```bash
make manifests
```

**Install the CRDs into the cluster**

```bash
make install
```

**Run the controllerr**

```bash
make run
```

</details>

```bash
make manifests install run
```

On a different terminal window, create PoolScaler resources

```bash
kubectl apply -k config/samples/
```

### 6. Testing

```bash
pipeci run --local
```

The system is able to run multiple pipelines asynchronously, but we recommend only run one at a time due to resource allocation limit.

### 7. Dashboards

Check service exposed by `kubectl get svc`. For quick reference:

- MinIO Console at `http://localhost:9090` (minio, minio123)
- RabbitMQ
  - Pipeline Queue at `http://localhost:15672` (guest, guest)
  - Job Queue at `http://localhost:15673` (guest, guest)

### 8. Delete the minikube cluster

The fastest way to clean up all resources is to delete the minikube cluster.

```bash
minikube delete --all --purge
```
