---
sidebar_position: 3
title: Running Parallel Agents
description: Learn to run multiple agents simultaneously on different feature branches
---

# Running Parallel Agents

In this tutorial, we will run multiple agents in parallel, each working on a different feature branch. By the end, you will understand how to maximize productivity by delegating multiple tasks simultaneously while maintaining complete isolation between them.

This tutorial takes approximately 30-40 minutes to complete.

## Prerequisites

Before starting, ensure you have:

- Completed the [Getting Started](./getting-started) tutorial
- Completed the [Your First Coding Task](./first-coding-task) tutorial
- A git repository with multiple areas that could be worked on independently
- Claude Code authentication configured (`hjk auth claude`)

## What We Will Build

We will simulate a realistic development scenario: preparing for a release that requires three independent changes:

1. Adding a new API endpoint
2. Improving error handling in an existing module
3. Writing documentation for a feature

Each task will run in its own isolated container with its own git worktree.

## Step 1: Plan Your Parallel Tasks

Before spawning agents, identify tasks that can run independently. Good candidates for parallelization:

- Touch different files or modules
- Do not depend on each other's changes
- Have clear, self-contained requirements

Our three tasks meet these criteria. Let us prepare prompts for each.

**Task 1: API endpoint**
```
Add a GET /api/health endpoint that returns { status: "ok", timestamp: <current time> }.
Include a test in the API test suite.
```

**Task 2: Error handling**
```
Improve error handling in src/services/payment.js. Wrap external API calls
in try-catch blocks, add logging for failures, and return user-friendly
error messages. Add tests for error scenarios.
```

**Task 3: Documentation**
```
Write API documentation for the user endpoints in docs/api/users.md.
Include request/response examples for each endpoint. Follow the format
used in docs/api/products.md.
```

## Step 2: Launch the First Agent

Navigate to your repository:

```bash
cd ~/projects/my-app
```

Launch the first agent in detached mode so we can start others immediately:

```bash
hjk run feat/health-endpoint --agent claude -d "Add a GET /api/health endpoint that returns { status: ok, timestamp: <current time> }. Include a test in the API test suite."
```

Output:

```
Created instance abc123 for branch feat/health-endpoint
Created session happy-panda in instance abc123 (detached)
```

The `-d` flag starts the session detached. The agent is now working in the background.

## Step 3: Launch Additional Agents

Without waiting for the first agent, launch the others:

```bash
hjk run feat/payment-errors --agent claude -d "Improve error handling in src/services/payment.js. Wrap external API calls in try-catch blocks, add logging for failures, and return user-friendly error messages. Add tests for error scenarios."
```

```bash
hjk run docs/user-api --agent claude -d "Write API documentation for the user endpoints in docs/api/users.md. Include request/response examples for each endpoint. Follow the format used in docs/api/products.md."
```

You now have three agents running simultaneously, each in its own isolated environment.

## Step 4: Monitor All Agents

View all running instances:

```bash
hjk ps
```

Output:

```
BRANCH              STATUS   SESSIONS  CREATED
feat/health-endpoint  running  1         2m ago
feat/payment-errors   running  1         1m ago
docs/user-api         running  1         30s ago
```

Each instance has its own container and git worktree. Changes in one cannot affect the others.

## Step 5: Check Individual Progress

To see detailed session information for an instance:

```bash
hjk ps feat/health-endpoint
```

Output:

```
SESSION       TYPE    STATUS    CREATED   ACCESSED
happy-panda   claude  detached  3m ago    just now
```

View the logs to check progress:

```bash
hjk logs feat/health-endpoint happy-panda
```

To follow logs in real-time:

```bash
hjk logs feat/health-endpoint happy-panda -f
```

Press `Ctrl+C` to stop following.

## Step 6: Cycle Through Agents

You can attach to any agent to observe or interact:

```bash
hjk attach feat/health-endpoint
```

Watch the agent work, answer any questions it has, then detach:

```
Ctrl+B, then d
```

Move to the next agent:

```bash
hjk attach feat/payment-errors
```

This workflow lets you check on each agent periodically, provide guidance when needed, and let them work autonomously otherwise.

## Step 7: Handle Agent Questions

When an agent has a question, it will pause and wait for input. Check logs periodically to see if any agent is waiting:

```bash
hjk logs feat/payment-errors happy-panda -n 20
```

If you see a question at the end of the output, attach and respond:

```bash
hjk attach feat/payment-errors
```

Provide your answer, then detach if the agent can continue independently.

## Step 8: Review Completed Work

As agents finish their tasks, review the results. Check which branches have completed work:

```bash
hjk ps
```

For a finished agent, you might see it has exited or is idle. Attach to confirm:

```bash
hjk attach docs/user-api
```

If Claude indicates it has completed the task, review the changes:

```bash
# Detach first
# Ctrl+B, then d

# Check out the branch
git checkout docs/user-api

# Review the documentation
cat docs/api/users.md
```

## Step 9: Run Multiple Sessions per Branch

Sometimes you want multiple agents working on the same branch for different subtasks. Use the `--name` flag to create named sessions:

```bash
hjk run feat/payment-errors --agent claude -d --name error-handling "Implement the try-catch wrappers for external API calls"

hjk run feat/payment-errors --agent claude -d --name error-tests "Write tests for the error handling scenarios in payment.js"
```

Both sessions share the same git worktree, so they can see each other's changes. This is useful when:

- One agent implements while another writes tests
- You want different agents to tackle different parts of a large feature
- You need a shell session alongside an agent session

List sessions for the branch:

```bash
hjk ps feat/payment-errors
```

Output:

```
SESSION         TYPE    STATUS    CREATED   ACCESSED
error-handling  claude  detached  2m ago    1m ago
error-tests     claude  detached  1m ago    just now
```

:::warning
Multiple sessions sharing a worktree can create conflicts if they modify the same files. Use this pattern when tasks touch different files.
:::

## Step 10: Clean Up Completed Work

As you finish reviewing each feature, clean up the instances:

```bash
# Stop instances you're done with
hjk stop feat/health-endpoint
hjk stop docs/user-api

# Remove instances completely when work is merged
hjk rm feat/health-endpoint
hjk rm docs/user-api
```

Keep the payment errors instance running if you want to continue iterating:

```bash
hjk ps
```

Output:

```
BRANCH              STATUS   SESSIONS  CREATED
feat/payment-errors running  2         15m ago
```

## What We Learned

In this tutorial, we:

- Identified tasks suitable for parallelization
- Launched multiple agents in detached mode
- Monitored parallel progress using `hjk ps` and `hjk logs`
- Cycled through agents to provide guidance
- Ran multiple sessions on a single branch for subtask division
- Cleaned up completed instances

The key insight is that each instance is completely isolated. You can safely run as many agents as your machine can handle without worrying about interference. This transforms sequential development into parallel development.

## Performance Considerations

Each agent runs in its own VM-isolated container, which consumes resources. On a typical development machine:

- 2-4 parallel agents run comfortably
- More agents may slow down due to CPU and memory constraints
- Monitor system resources if running many agents

## Next Steps

Now that you can run parallel agents, explore these resources:

**Tutorials**
- [Building a Custom Image](./custom-image) - Pre-install dependencies for faster agent startup

**How-To Guides**
- [Manage Sessions](../how-to/manage-sessions) - Advanced session patterns
- [Stop and Clean Up Instances](../how-to/stop-cleanup) - Bulk operations for many instances

**Concepts**
- [Isolation Model](../explanation/isolation-model) - How containers ensure safety
- [Worktree Strategy](../explanation/worktree-strategy) - How branch isolation works
