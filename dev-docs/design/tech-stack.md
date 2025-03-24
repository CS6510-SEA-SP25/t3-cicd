# Tech stack

This document contains the tech stack is used for the CI/CD system. Trade-offs for each technology choice are also discussed.

| Component            | Technology Choices |
| -------------------- | ------------------ |
| **API**              | Go                 |
| **Database**         | MySQL              |
| **CLI**              | Go Cobra           |
| **Containers**       | Docker, Kubernetes |
| **Artifact storage** | MinIO              |
| **Message Queue**    | RabbitMQ           |

We have considerations for technology choices:

#### API: Golang

Concurrency is one important issue for the CI/CD system, as the system should be able to handle multiple jobs at the same time. Goroutines are lightweight threads of execution that make it effortless to write concurrent code, while channels provide a safe way for these goroutines to communicate and coordinate.

#### Database: PostgreSQL vs MySQL

One major difference is that MySQL suits read-heavy operations and PostgreSQL suits write-heavy operations.
We expect the CI/CD system involves a lot of read queries, so MySQL seems like a better fit.
PostgreSQL has a built-in fulltext search feature, which might be beneficial if we want to use it to analyze logs.
With MySQL, this can be solved by using Elasticsearch.

#### Job Execution: Docker

We choose Docker for its popularity and community support.

#### Cloud services: AWS

The team is familiar with AWS services, which shorten the learning curve.

#### CLI application: Go Cobra

Go compiles to executables, which is perfect for creating CLI application. We can also release open-source package manager like HomeBrew with GoReleaser with basic setup.
