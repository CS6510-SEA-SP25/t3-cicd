# Tech stack

This document contains the tech stack is used for the CI/CD system. Trade-offs for each technology choice are also discussed.

| Component          | Technology Choices |
| ------------------ | ------------------ |
| **API**            | Spring Boot (Java) |
| **Database**       | MySQL              |
| **Job Execution**  | Docker             |
| **Cloud Services** | AWS                |
| **CLI**            | Python             |

We have considerations for technology choices:

#### API: We consider Java Spring Boot and Node.js.

Concurrency is one important issue for the CI/CD system, as the system should be able to handle multiple jobs at the same time.
For Java, each request gets its own thread, while JS requests go into the same thread and are handled with event loop.
At scale, Java can use more threads, thus can handle huge work load and multiple long-running tasks at once.
JS can also handle CPU heavy tasks like compile code and running tests, but with [worker threads](https://nodejs.org/api/worker_threads.html),
which incur additional learning curve and extra setup.

#### Database: PostgreSQL vs MySQL

One major difference is that MySQL suits read-heavy operations and PostgreSQL suits write-heavy operations.
We expect the CI/CD system involves a lot of read queries, so MySQL seems like a better fit.
PostgreSQL has a built-in fulltext search feature, which might be beneficial if we want to use it to analyze logs.
With MySQL, this can be solved by using Elasticsearch.

#### Job Execution: Docker

We choose Docker for its popularity and community support.

#### Cloud services: AWS

The team is familiar with AWS services, which shorten the learning curve.

#### CLI application: Python

We choose Python for the ease of development. Ideally, Go would be a great idea due to its performance and startup time, and Go doesn't require runtime installed like Python and Java.

Java is too verbose, so we cross it out.

Using Go would incur learning curve. Hence, we choose Python for now.