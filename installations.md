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

```bash
pipeci help
```

### Backend

There are 3 ways to run local for different use cases.

#### 1. Go run

**Prerequisites**: Please set GITHUB_TOKEN at your local env by running `export GITHUB_TOKEN=...`

The easiest and fastest way is to run go api.

```bash
cd backend && go run .
```

Use this for dev and debug.

#### 2. Minikube

We setup a minikube cluster. Please follow the instructions at root directory.

Start the minikube cluster

```bash
minikube start
```

Create `k8s/github-secret.yaml` from this template

```
apiVersion: v1
kind: Secret
metadata:
  name: github-secret
type: Opaque
data:
  token: [value achieved by running:"echo -n $GITHUB_TOKEN | base64"]
```

Deploy to minikube

```bash
❯ ./deploy.sh
```

To check for status, please run

```bash
❯ kubectl get pods
```

Run this on a different terminal window, or consider running it in background.

```bash
❯ minikube tunnel
```

Check for deployments

```bash
❯ kubectl get deployment
NAME             READY   UP-TO-DATE   AVAILABLE   AGE
cicd-api         1/1     1            1           169m
minio-operator   2/2     2            2           169m
mysql            1/1     1            1           169m
```

And check for k8s service

```bash
❯ kubectl get svc
NAME               TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
cicd-api-service   LoadBalancer   10.100.13.91     <pending>     8080:32006/TCP   169m
kubernetes         ClusterIP      10.96.0.1        <none>        443/TCP          169m
minio              LoadBalancer   10.96.130.139    <pending>     80:32564/TCP     168m
myminio-console    LoadBalancer   10.111.6.120     <pending>     9090:31710/TCP   168m
myminio-hl         ClusterIP      None             <none>        9000/TCP         168m
mysql              ClusterIP      10.100.47.5      <none>        3306/TCP         169m
operator           ClusterIP      10.109.211.246   <none>        4221/TCP         169m
sts                ClusterIP      10.105.151.58    <none>        4223/TCP         169m
```

To verify the API, please check the endpoint

```
http://REPLACE_WITH_EXTERNAL_IP:8080
```

For our application,

- `http://localhost:8080` to ping the server.
- `http://localhost:9090` to access the MinIO Console.
