# Headjack Base Image
#
# A complete development environment for CLI-based coding agents.
# See docs/designs/base-image.md for full specification.

FROM ubuntu:24.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Set locale and timezone
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8
ENV TZ=UTC

# Default editor
ENV EDITOR=vim

# =============================================================================
# Base System Setup
# =============================================================================

RUN apt-get update && apt-get install -y \
    # Locale support
    locales \
    # Timezone
    tzdata \
    # systemd (required for multi-service environments)
    systemd \
    systemd-sysv \
    # iptables (required for Docker-in-Docker)
    iptables \
    # SSL certificates
    ca-certificates \
    # Basic utilities needed for subsequent installations
    curl \
    wget \
    gnupg \
    lsb-release \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# Configure locale
RUN locale-gen en_US.UTF-8

# =============================================================================
# iptables-legacy Workaround (ADR-002)
# Required for Docker-in-Docker in Apple Containerization Framework
# =============================================================================

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

# =============================================================================
# Common Tooling
# =============================================================================

RUN apt-get update && apt-get install -y \
    # Version control
    git \
    git-lfs \
    # Network & data utilities
    jq \
    # File & text processing
    ripgrep \
    fd-find \
    fzf \
    tree \
    less \
    # System utilities
    htop \
    vim \
    openssh-client \
    zip \
    unzip \
    make \
    build-essential \
    # Additional build dependencies for version managers
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    libxml2-dev \
    libxmlsec1-dev \
    libffi-dev \
    liblzma-dev \
    && rm -rf /var/lib/apt/lists/*

# =============================================================================
# GitHub CLI (gh)
# =============================================================================

RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg && \
    chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null && \
    apt-get update && \
    apt-get install -y gh && \
    rm -rf /var/lib/apt/lists/*

# =============================================================================
# yq (YAML processor)
# =============================================================================

# TARGETARCH is automatically set by Docker buildx for multi-platform builds
ARG TARGETARCH
RUN curl -fsSL "https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${TARGETARCH}" -o /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

# =============================================================================
# Docker CE
# =============================================================================

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update && \
    apt-get install -y \
    docker-ce \
    docker-ce-cli \
    containerd.io \
    docker-buildx-plugin \
    docker-compose-plugin \
    && rm -rf /var/lib/apt/lists/*

# Enable Docker to start with systemd
RUN systemctl enable docker

# =============================================================================
# Node.js (required for Agent CLIs)
# =============================================================================

RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

# =============================================================================
# Agent CLIs
# =============================================================================

RUN npm install -g @anthropic-ai/claude-code && \
    npm install -g @google/gemini-cli && \
    npm install -g @openai/codex

# =============================================================================
# Create non-root user
# =============================================================================

ARG USERNAME=developer
ARG USER_UID=1000
ARG USER_GID=1000

# Check if user/group with these IDs already exist, reuse or create as needed
RUN if getent passwd $USER_UID > /dev/null 2>&1; then \
        # User with this UID exists, get their name and rename if needed
        existing_user=$(getent passwd $USER_UID | cut -d: -f1); \
        if [ "$existing_user" != "$USERNAME" ]; then \
            usermod -l $USERNAME -d /home/$USERNAME -m $existing_user; \
        fi; \
    else \
        # Create group if needed
        if ! getent group $USER_GID > /dev/null 2>&1; then \
            groupadd --gid $USER_GID $USERNAME; \
        fi; \
        # Create user
        useradd --uid $USER_UID --gid $USER_GID -m $USERNAME; \
    fi && \
    usermod -aG docker $USERNAME

# =============================================================================
# Version Managers (installed as user)
# =============================================================================

USER $USERNAME
WORKDIR /home/$USERNAME

# pyenv
RUN git clone https://github.com/pyenv/pyenv.git ~/.pyenv

# nodenv
RUN git clone https://github.com/nodenv/nodenv.git ~/.nodenv && \
    git clone https://github.com/nodenv/node-build.git ~/.nodenv/plugins/node-build

# goenv
RUN git clone https://github.com/go-nv/goenv.git ~/.goenv

# rustup
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path

# =============================================================================
# Shell Configuration
# =============================================================================

RUN { \
    echo ''; \
    echo '# pyenv'; \
    echo 'export PYENV_ROOT="$HOME/.pyenv"'; \
    echo 'export PATH="$PYENV_ROOT/bin:$PATH"'; \
    echo 'eval "$(pyenv init -)"'; \
    echo ''; \
    echo '# nodenv'; \
    echo 'export PATH="$HOME/.nodenv/bin:$PATH"'; \
    echo 'eval "$(nodenv init -)"'; \
    echo ''; \
    echo '# goenv'; \
    echo 'export GOENV_ROOT="$HOME/.goenv"'; \
    echo 'export PATH="$GOENV_ROOT/bin:$PATH"'; \
    echo 'eval "$(goenv init -)"'; \
    echo ''; \
    echo '# rustup'; \
    echo 'source "$HOME/.cargo/env"'; \
    echo ''; \
    echo '# fd-find is installed as fdfind on Ubuntu'; \
    echo 'alias fd=fdfind'; \
    } >> ~/.bashrc

# =============================================================================
# systemd Configuration
# =============================================================================

USER root

# Clean up systemd services that don't make sense in a container
RUN rm -f /lib/systemd/system/multi-user.target.wants/* \
    /etc/systemd/system/*.wants/* \
    /lib/systemd/system/local-fs.target.wants/* \
    /lib/systemd/system/sockets.target.wants/*udev* \
    /lib/systemd/system/sockets.target.wants/*initctl* \
    /lib/systemd/system/sysinit.target.wants/systemd-tmpfiles-setup* \
    /lib/systemd/system/systemd-update-utmp*

# Use systemd as init
STOPSIGNAL SIGRTMIN+3

# Volume for cgroups (required for systemd)
VOLUME ["/sys/fs/cgroup"]

# Default to the developer user
USER $USERNAME
WORKDIR /home/$USERNAME

# systemd as init system
CMD ["/lib/systemd/systemd"]
