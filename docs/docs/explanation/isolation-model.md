---
sidebar_position: 2
title: Isolation Model
description: Why VM-based container isolation matters for running AI agents safely
---

# Isolation Model

Running autonomous AI coding agents requires strong isolation. These agents execute arbitrary code, install packages, modify system configurations, and interact with external services. Headjack uses VM-based container isolation to ensure that agents cannot escape their sandbox, interfere with each other, or compromise the host system.

## The Threat Model

When you run a coding agent, you're giving it significant capabilities:

- **Code execution**: Agents run shell commands, compile code, and execute tests
- **File system access**: Agents read and write files throughout the workspace
- **Network access**: Agents fetch packages, interact with APIs, and communicate with services
- **Package installation**: Agents install dependencies, sometimes with native extensions

A malicious or confused agent could attempt to:

- Access files outside its workspace
- Read secrets from other processes or the host
- Consume excessive resources
- Install persistent malware
- Escalate privileges

Even well-intentioned agents can cause problems through mistakes: deleting important files, modifying system configurations, or installing conflicting packages. Isolation contains the blast radius.

## Traditional Container Isolation

Standard container runtimes like Docker use Linux namespaces and cgroups to isolate processes. This provides:

- **Filesystem isolation**: Each container has its own root filesystem
- **Process isolation**: Containers see only their own processes
- **Network isolation**: Containers have their own network namespace
- **Resource limits**: Cgroups constrain CPU and memory usage

However, namespace-based isolation has limitations:

- **Shared kernel**: All containers share the host's Linux kernel. Kernel vulnerabilities affect all containers.
- **Privileged operations**: Some workloads (like Docker-in-Docker) require privileged mode, which weakens isolation significantly.
- **Container escapes**: History shows that container escapes are possible, especially with privileged containers.

For running untrusted code from AI agents, namespace isolation alone is insufficient.

## VM-Based Isolation

Headjack uses the **Apple Containerization Framework**, which runs each container in its own lightweight virtual machine. This provides a stronger isolation boundary:

```
+-----------------------------------------------------------+
|                     macOS Host                            |
|  +--------------------+  +--------------------+           |
|  |       VM 1         |  |       VM 2         |           |
|  |  +--------------+  |  |  +--------------+  |           |
|  |  |  Container   |  |  |  |  Container   |  |           |
|  |  |   Agent A    |  |  |  |   Agent B    |  |           |
|  |  +--------------+  |  |  +--------------+  |           |
|  |   Linux Kernel     |  |   Linux Kernel     |           |
|  +--------------------+  +--------------------+           |
|                                                           |
|                     Hypervisor                            |
+-----------------------------------------------------------+
```

Each agent runs in a complete Linux environment with its own kernel. The hypervisor (Apple's Virtualization.framework) enforces the boundary at the hardware level. This provides:

- **Kernel isolation**: Each VM has its own Linux kernel. A kernel vulnerability in one VM cannot affect others.
- **Hardware-level separation**: The hypervisor uses hardware virtualization features (ARM's virtualization extensions) to enforce boundaries.
- **No shared state**: VMs share nothing by default. File sharing requires explicit mounts.
- **Safe privileged operations**: Docker-in-Docker works inside a VM without compromising host security.

## Why Not Just Docker?

Docker is excellent for many containerization use cases, but it optimizes for different goals than Headjack requires:

| Concern | Docker | Apple Containerization |
|---------|--------|----------------------|
| Isolation boundary | Shared kernel (namespaces) | Hardware (hypervisor) |
| Docker-in-Docker | Requires privileged mode | Works without host compromise |
| Startup time | Sub-second | ~1 second |
| Memory overhead | Minimal (shared kernel) | Per-VM kernel (~50MB) |
| Primary use case | Application packaging | Isolated development |

For server workloads where you control what runs in containers, Docker's isolation model is appropriate. For running autonomous agents that execute arbitrary code, the stronger guarantees of VM-based isolation are worth the overhead.

## The Performance Trade-off

VM-based isolation has costs:

- **Memory**: Each VM runs its own kernel, consuming approximately 50MB of overhead per instance
- **Startup**: VM boot takes about a second (still fast, but not instantaneous)
- **No memory ballooning**: VMs have fixed memory allocation (no dynamic sharing with host)

For Headjack's use case, these trade-offs are acceptable:

- **Memory**: Running 10 concurrent agents requires perhaps 500MB of kernel overhead. On a machine with 16GB+ RAM, this is negligible compared to the agents themselves.
- **Startup**: A one-second startup is imperceptible when you're about to work with an agent for minutes or hours.
- **Fixed memory**: Development machines typically have ample RAM, and memory contention is rarely the bottleneck.

The security benefits outweigh these costs for the autonomous agent use case.

## What Gets Isolated

When Headjack creates an instance, the following are isolated per-agent:

| Component | Isolation Level |
|-----------|----------------|
| Kernel | Separate Linux kernel per VM |
| Filesystem | Container image + mounted worktree only |
| Processes | Only visible within the container |
| Network | Separate network namespace with NAT |
| Users | Container runs as non-root by default |
| Packages | Installed only in container |

The worktree mount is the controlled bridge between host and container. The agent can read and write files in the workspace, but nothing else from the host filesystem.

## Shared Resources

Some resources are intentionally shared or passed through:

- **Git repository**: Via the worktree mount at `/workspace`
- **Authentication tokens**: Injected as environment variables at session start
- **Network access**: Containers can reach the internet (with NAT)
- **Time**: VMs synchronize with host time

These represent the minimal surface area needed for agents to function. Each sharing decision is explicit and intentional.

## Isolation vs. Convenience

Strong isolation sometimes conflicts with convenience:

- **No GPG forwarding** (yet): GPG agent sockets don't cross VM boundaries. Commit signing requires workarounds. See [ADR-005](../decisions/adr-005-no-gpg-support).
- **No clipboard sharing**: Copy/paste between agent output and host requires manual transfer.
- **No host Docker socket**: Agents can run Docker inside their VM, but cannot access the host's Docker daemon.

These limitations are features, not bugs. Each represents a potential attack surface that Headjack intentionally avoids. The workarounds exist for users who need these capabilities and understand the trade-offs.

## Trust Boundaries

Understanding trust boundaries helps reason about security:

```
Untrusted
    |
    v
+---------------------------------------------+
|  Agent output, generated code               |
+---------------------------------------------+
|  Agent CLI (claude, gemini, codex)          |
+---------------------------------------------+
|  Container environment                      |
+---------------------------------------------+
|  Linux kernel (in VM)                       |
+=============================================+  <-- Hypervisor boundary
|  macOS Virtualization.framework             |
+---------------------------------------------+
|  macOS kernel                               |
+---------------------------------------------+
    |
    v
Trusted
```

Everything above the hypervisor boundary is assumed potentially compromised. The hypervisor boundary is enforced by hardware. This model accepts that agents might be tricked into running malicious code, but ensures that code cannot escape the VM.

## Related

- [ADR-002: Apple Containerization Framework](../decisions/adr-002-apple-containerization) - Why this runtime was chosen
- [Architecture Overview](./architecture) - How isolation fits into the broader architecture
- [Image Customization](./image-customization) - Building custom container images
