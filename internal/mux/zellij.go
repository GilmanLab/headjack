package mux

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/jmgilman/headjack/internal/exec"
)

// zellij implements Multiplexer using the Zellij terminal multiplexer.
type zellij struct {
	exec exec.Executor
}

// NewZellij creates a Multiplexer using Zellij CLI.
func NewZellij(e exec.Executor) *zellij {
	return &zellij{exec: e}
}

func (z *zellij) CreateSession(ctx context.Context, opts CreateSessionOpts) (*Session, error) {
	// Check if session already exists
	sessions, err := z.ListSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("check existing sessions: %w", err)
	}

	for _, s := range sessions {
		if s.Name == opts.Name {
			return nil, ErrSessionExists
		}
	}

	// Build command arguments
	// zellij --session <name> [options...]
	args := []string{"--session", opts.Name}

	// Add working directory if specified
	if opts.Cwd != "" {
		args = append(args, "--cwd", opts.Cwd)
	}

	// If a command is specified, we need to create the session in detached mode
	// and then run the command. Zellij doesn't have a direct "run command in new session" option,
	// so we create the session and it will start with the default shell.
	// The command will be run by the caller via container exec.

	// For headjack, sessions are created inside containers, so we start zellij
	// in the background/detached mode. The session creation happens when zellij starts.

	// Create session - zellij will create it if it doesn't exist when we attach
	// But for background sessions, we need to start zellij in a way that it detaches
	// Unfortunately, zellij doesn't have a native "create and detach" command.
	// We'll create it by starting zellij and immediately detaching.

	// For now, we just prepare the session info. The actual session creation
	// happens when AttachSession is called (zellij creates if it doesn't exist).
	// This matches zellij's behavior where attach creates if needed.

	return &Session{
		ID:        opts.Name, // Zellij uses session name as ID
		Name:      opts.Name,
		CreatedAt: time.Now(),
	}, nil
}

func (z *zellij) AttachSession(ctx context.Context, sessionName string) error {
	// zellij attach <session-name> or zellij --session <name> (creates if not exists)
	args := []string{"attach", sessionName, "--create"}

	stdinFd := int(os.Stdin.Fd())

	// Check if stdin is a terminal
	if !term.IsTerminal(stdinFd) {
		// Fall back to non-interactive mode
		_, err := z.exec.Run(ctx, exec.RunOptions{
			Name:   "zellij",
			Args:   args,
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		})
		if err != nil {
			return fmt.Errorf("%w: %v", ErrAttachFailed, err)
		}
		return nil
	}

	// Put terminal in raw mode for proper TTY handling
	oldState, err := term.MakeRaw(stdinFd)
	if err != nil {
		return fmt.Errorf("set terminal raw mode: %w", err)
	}
	defer term.Restore(stdinFd, oldState)

	// Handle window resize signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	defer signal.Stop(sigCh)

	// Run zellij with stdio attached
	_, err = z.exec.Run(ctx, exec.RunOptions{
		Name:   "zellij",
		Args:   args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAttachFailed, err)
	}

	return nil
}

func (z *zellij) ListSessions(ctx context.Context) ([]Session, error) {
	// zellij list-sessions
	result, err := z.exec.Run(ctx, exec.RunOptions{
		Name: "zellij",
		Args: []string{"list-sessions"},
	})
	if err != nil {
		// If zellij exits with error but has no sessions, that's ok
		stderr := string(result.Stderr)
		if strings.Contains(stderr, "No active") || result.ExitCode == 0 {
			return []Session{}, nil
		}
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	// Parse output - each line is a session name
	// Format: "session-name [Created ...] (current)" or just "session-name"
	output := strings.TrimSpace(string(result.Stdout))
	if output == "" {
		return []Session{}, nil
	}

	lines := strings.Split(output, "\n")
	sessions := make([]Session, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract session name (first word before any brackets or parentheses)
		name := line
		if idx := strings.IndexAny(line, " \t[("); idx > 0 {
			name = line[:idx]
		}

		sessions = append(sessions, Session{
			ID:   name,
			Name: name,
			// CreatedAt is not reliably available from list output
		})
	}

	return sessions, nil
}

func (z *zellij) KillSession(ctx context.Context, sessionName string) error {
	// zellij kill-session <session-name>
	result, err := z.exec.Run(ctx, exec.RunOptions{
		Name: "zellij",
		Args: []string{"kill-session", sessionName},
	})
	if err != nil {
		stderr := string(result.Stderr)
		if strings.Contains(stderr, "not found") || strings.Contains(stderr, "No session") {
			return ErrSessionNotFound
		}
		return fmt.Errorf("kill session: %w", err)
	}

	return nil
}
