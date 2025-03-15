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

Add the `init.sql` file to Config Map

```bash
kubectl create configmap mysql-init-config --from-file=backend/db/init.sql
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
❯ kubectl apply -f k8s
```

To check for status, please run

```bash
❯ kubectl get pods
```

Once all pods runs successfully, we need to expose the k8s deployment at localhost:8080 by using LoadBalancer

Run this on a different terminal window, or consider running it in background.

```bash
❯ minikube tunnel
```

Check for deployments

```bash
❯ kubectl get deployment
NAME       READY   UP-TO-DATE   AVAILABLE   AGE
cicd-api   1/1     1            1           10h
mysql      1/1     1            1           10h
```

Expose at port 8080

```bash
❯ kubectl expose deployment cicd-api --type=LoadBalancer --port=8080
```

And check for the external IP

```bash
kubectl get svc
```

To verify the API, please check the endpoint

```
http://REPLACE_WITH_EXTERNAL_IP:8080
```

For our application, go to `http://localhost:8080` to ping the server.

#### 3. Docker compose

Alternatively, you can update `docker-compose.yaml` and use it a your own risk as we won't maintain this file going forward.

Run the backend server and MySQL database on with docker-compose. At the root directory:

```bash
❯ docker-compose up -d
```

To verify the backend is working:

```bash
❯ pipeci report --local
pipeci: Using input configuration file at .pipelines/pipeline.yaml
pipeci: Pipeline Details:
pipeci:   Commit Hash: fc632d88dfbe004cda2153f3244f9272a8f4d893
pipeci:   Name: maven_project_1
pipeci:   Repository: https://github.com/CS6510-SEA-SP25/hw3-minh160302.git
pipeci:   Pipeline ID: 1
pipeci:   Status: SUCCESS
pipeci:   Start Time: 2025-02-24T01:16:15Z
pipeci:   End Time: 2025-02-24T01:17:07Z
pipeci:   IP Address: 0.0.0.0
pipeci:   Stage Order: verify
pipeci: ----------------------------------------
```
