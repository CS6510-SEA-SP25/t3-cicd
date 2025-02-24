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

Run the backend server and MySQL database on with docker-compose. At the root directory:

```bash
docker-compose up -d
docker-compose up -d
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
