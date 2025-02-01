# Team Processes

Repository configuration:

- one single repository
- Repository organization

  ```
  ├── README.md
  ├── cli                 # CLI application written in Python
  ├── backend             # Main backend code written in Java Spring Boot
  └── .github             # Github Actions CI/CD
  ```

  The folder structure is subject to change during development.

- CI/CD setup: typical GitHub Actions workflows
- Testing coverage: Jacoco (Java) and [Coverage.py](https://coverage.readthedocs.io/en/7.6.10/) (Python)
- Code style: Checkstyle (Java) and [autopep8](https://pypi.org/project/autopep8/) (Python)
- Static Analysis: SpotBugs (Java) and [Pylint](https://pylint.readthedocs.io/en/stable/) (Python)
