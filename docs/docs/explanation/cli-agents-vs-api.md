---
sidebar_position: 3
title: CLI Agents vs API
description: Why Headjack uses CLI agents with subscriptions instead of API-based agents
---

# CLI Agents vs API

Headjack orchestrates CLI-based coding agents rather than calling LLM APIs directly. This is a deliberate choice driven primarily by economics, but also by practical considerations around tooling and maintenance burden.

## The Economic Reality

API-based LLM access uses token-based pricing. You pay for every input token (your prompt, context, file contents) and every output token (the model's response). For interactive chat, this can be reasonable. For autonomous coding agents, costs escalate quickly.

Consider what an autonomous agent does during a typical task:

1. Reads the task description and codebase context (input tokens)
2. Generates a plan (output tokens)
3. Reads files to understand the codebase (input tokens)
4. Generates code changes (output tokens)
5. Runs tests and reads results (input tokens)
6. Iterates based on failures (more input and output tokens)
7. Refines until satisfied (repeated cycles)

A single task might involve reading dozens of files, generating hundreds of lines of code, and iterating through multiple test cycles. At current API prices, a complex task could easily cost $5-20 in token fees.

CLI agents with subscription pricing fundamentally change this equation:

| Model | Pricing Model | Typical Cost |
|-------|---------------|--------------|
| API (Claude, GPT-4) | ~$15/million input tokens, ~$75/million output tokens | $5-50+ per complex task |
| Claude Code (Max) | $100-200/month subscription | Unlimited* |
| Gemini CLI | Included with Gemini subscription | ~$20/month |

*Subject to fair use policies

For power users running agents frequently, subscriptions offer 10-100x cost savings compared to API access. Headjack exists precisely to maximize the value of these subscriptions by enabling multiple concurrent agents.

## Pre-Built Capabilities

Beyond economics, CLI agents come with significant built-in capabilities that would be expensive to replicate:

### Tool Integration

Claude Code, Gemini CLI, and Codex all include:

- File reading and writing
- Command execution
- Search tools (grep, find, ripgrep)
- Git operations
- Web browsing (in some cases)

Building equivalent tool integration for a custom API-based agent requires significant development effort: defining tool schemas, implementing execution logic, handling errors, managing permissions, and iterating based on real-world usage.

### Prompt Engineering

CLI agents incorporate extensive prompt engineering developed through millions of user interactions:

- System prompts that establish effective coding behavior
- Tool-use patterns that balance autonomy with safety
- Context management strategies that work within token limits
- Output formatting that's readable in terminal contexts

This prompt engineering represents substantial R&D investment that Headjack gets "for free" by using the official CLIs.

### Continuous Improvement

As Anthropic, Google, and OpenAI improve their agents, Headjack benefits automatically. New model capabilities, better tool integration, and improved prompts flow through CLI updates without any changes to Headjack itself.

## The Process Model

Headjack's architecture treats agents as Unix processes rather than API clients:

```
+-----------------------------------------------------------+
|  Traditional API-based approach:                          |
|                                                           |
|  +---------------+      +---------------+                 |
|  | Orchestrator  | ---> |   LLM API     |                 |
|  |   (Python)    | <--- |   (HTTP)      |                 |
|  +---------------+      +---------------+                 |
|        |                                                  |
|        v                                                  |
|  [Tool execution, context management, etc.]               |
+-----------------------------------------------------------+

+-----------------------------------------------------------+
|  Headjack's process-based approach:                       |
|                                                           |
|  +---------------+      +---------------+                 |
|  |   Headjack    | ---> |  CLI Agent    | ---> [LLM API] |
|  |   (spawn)     |      |  (process)    |                 |
|  +---------------+      +---------------+                 |
|                              |                            |
|                              v                            |
|                     [Tools built into CLI]                |
+-----------------------------------------------------------+
```

This process model has several advantages:

- **Isolation**: Each agent process has its own memory, environment, and state
- **Simplicity**: Headjack manages processes and sessions, not API orchestration
- **Debuggability**: Standard Unix tools (ps, top, logs) work as expected
- **Reliability**: Process crashes are isolated; Headjack doesn't need complex retry logic

## What We Give Up

The process-based approach has trade-offs:

### Less Control

API-based agents allow fine-grained control over:
- System prompts (though CLI agents accept some customization)
- Tool definitions (add custom tools specific to your workflow)
- Temperature and other generation parameters
- Output format (streaming, structured JSON, etc.)

With CLI agents, you accept the vendor's choices about these parameters. For most coding tasks, their choices are good. For specialized use cases, the lack of control can be limiting.

### Vendor Lock-in

Headjack depends on the continued existence and compatibility of CLI agents. If Anthropic discontinues Claude Code or changes its interface significantly, Headjack must adapt.

This risk is mitigated by supporting multiple agents (Claude, Gemini, Codex). If one becomes unavailable or too expensive, users can switch to another.

### Authentication Complexity

CLI agents have their own authentication flows (OAuth, API keys, device codes). Headjack must handle each agent's auth mechanism separately:

- Claude Code: OAuth token via `claude setup-token`
- Gemini CLI: Google OAuth via browser
- Codex: OpenAI OAuth via browser

API-based agents have simpler auth: just provide an API key. The multi-step OAuth flows required by CLI agents add complexity to Headjack's auth system.

## Why Not Both?

A future version of Headjack could support API-based agents alongside CLI agents. The session abstraction is flexible enough to run any process:

```bash
# Current: CLI agents
hjk run feature-branch --agent claude

# Future: API-based agent (hypothetical)
hjk run feature-branch --agent api --model claude-sonnet --api-key $KEY
```

This remains out of scope for now because:

1. The economics strongly favor subscription-based CLI agents for power users
2. Building robust API-based agent tooling is a significant undertaking
3. The market is evolving rapidly; premature investment might be wasted

## The Market Context

The CLI agent market is maturing rapidly. When Headjack started, Claude Code was the only serious option. Now we have:

- **Claude Code** (Anthropic): The pioneer, most mature
- **Gemini CLI** (Google): Strong integration with Google services
- **Codex** (OpenAI): Leverages GPT-4 and ChatGPT ecosystem

Competition drives improvement. Each vendor has incentives to make their CLI agent better, faster, and more capable. Headjack benefits from this competition by remaining agent-agnostic.

## Summary

Headjack uses CLI agents because:

1. **Subscription pricing** offers 10-100x cost savings over API pricing for heavy users
2. **Pre-built tools** eliminate the need to implement file operations, command execution, etc.
3. **Continuous improvement** means Headjack benefits from vendor R&D automatically
4. **Process model** provides natural isolation and simplifies Headjack's architecture

The trade-off is less control over agent internals, but for the autonomous coding use case, the benefits strongly outweigh the limitations.

## Related

- [ADR-004: CLI-Based Agents](../decisions/adr-004-cli-agents) - The formal decision record
- [Authentication](./authentication) - How Headjack handles agent auth
- [Architecture Overview](./architecture) - How agents fit into the broader system
