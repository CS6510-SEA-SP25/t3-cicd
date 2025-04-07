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

For this version, the only way to run the project on local is using minikube. If you prefer run as Go project, please see the [`async`](https://github.com/CS6510-SEA-SP25/t3-cicd/tree/async) branch.

Below is the setup for the minikube cluster. Please follow the instructions at root directory.

### 1. Start the minikube cluster

```bash
minikube start
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

We are using [Aiven](https://aiven.io/mysql) as the managed service for MySQL. As the hosted database is only accessible with SSL, we need a CA certificate file.

Create a `ca.pem` file at root directory.

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


Deploy the controller to the cluster

```bash
cd operator/hpa && make deploy IMG=minh160302/pooloperator-operator
```

Create a PoolScaler

```bash
kubectl apply -k config/samples/
```
