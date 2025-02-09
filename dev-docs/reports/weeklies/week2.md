# Week 2

# Completed tasks

| Task                                              | Weight |
| ------------------------------------------------- | ------ |
| [First draft High-level Architecture documentation](https://github.com/CS6510-SEA-SP25/t3-cicd/issues/1) | 3      |

# Carry over tasks

This week we focus on the CLI implementation.


| Task                                         | Weight | Assignee    |
| -------------------------------------------- | ------ | ----------- |
| [Review High-level Architecture documentation](https://github.com/CS6510-SEA-SP25/t3-cicd/issues/2) | 3      | Minh Nguyen |

# New tasks

| Task                               | Weight | Assignee    |
| ---------------------------------- | ------ | ----------- |
| Define Pipeline configuration file | 1      | Minh Nguyen |
| Configuration file parsing         | 5      | Minh Nguyen |
| Implement functioning CLI commands | 5      | Minh Nguyen |
| Polish CLI documentations/reports  | 3      | Minh Nguyen |

# What worked this week?

> In this section list part of the team's process that you believe worked well. "Worked Well" means helped the team be more efficient and/or effective. Try to explain **why** these actions worked well.

# What did not work this week?

> In this section list part of the team's process that you believe did **not** work well. "Not Worked Well" means that the team found these actions to **not have a good effect** on the team's effectiveness. Try to explain **why** these actions did not work well.

# Design updates

- Backend Java -> Go
- Goroutines for concurrency in local execution
- Remote: several options. Consider tradeoffs:
    - k8s Jobs
    - Argo
    - MQ, ...