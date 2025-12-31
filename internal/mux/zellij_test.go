package mux

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmgilman/headjack/internal/exec"
	"github.com/jmgilman/headjack/internal/exec/mocks"
)

func TestNewZellij(t *testing.T) {
	mockExec := &mocks.ExecutorMock{}
	z := NewZellij(mockExec)

	require.NotNil(t, z)
	assert.Equal(t, mockExec, z.exec)
}

func TestZellij_CreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("creates session successfully", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				// list-sessions call returns empty
				assert.Equal(t, "zellij", opts.Name)
				assert.Equal(t, []string{"list-sessions"}, opts.Args)

				return &exec.Result{
					Stdout:   []byte(""),
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		session, err := z.CreateSession(ctx, CreateSessionOpts{
			Name: "test-session",
		})

		require.NoError(t, err)
		assert.Equal(t, "test-session", session.ID)
		assert.Equal(t, "test-session", session.Name)
		assert.False(t, session.CreatedAt.IsZero())
	})

	t.Run("returns ErrSessionExists when session exists", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				// list-sessions returns existing session
				return &exec.Result{
					Stdout:   []byte("test-session [Created ...]\nother-session\n"),
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		_, err := z.CreateSession(ctx, CreateSessionOpts{
			Name: "test-session",
		})

		assert.ErrorIs(t, err, ErrSessionExists)
	})

	t.Run("propagates list sessions error", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stderr:   []byte("zellij error"),
					ExitCode: 1,
				}, errors.New("exit code 1")
			},
		}

		z := NewZellij(mockExec)
		_, err := z.CreateSession(ctx, CreateSessionOpts{
			Name: "test-session",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "check existing sessions")
	})
}

func TestZellij_ListSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("returns empty list when no sessions", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				assert.Equal(t, "zellij", opts.Name)
				assert.Equal(t, []string{"list-sessions"}, opts.Args)

				return &exec.Result{
					Stdout:   []byte(""),
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		sessions, err := z.ListSessions(ctx)

		require.NoError(t, err)
		assert.Empty(t, sessions)
	})

	t.Run("parses single session", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stdout:   []byte("my-session\n"),
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		sessions, err := z.ListSessions(ctx)

		require.NoError(t, err)
		require.Len(t, sessions, 1)
		assert.Equal(t, "my-session", sessions[0].ID)
		assert.Equal(t, "my-session", sessions[0].Name)
	})

	t.Run("parses multiple sessions", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stdout:   []byte("session-1\nsession-2\nsession-3\n"),
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		sessions, err := z.ListSessions(ctx)

		require.NoError(t, err)
		require.Len(t, sessions, 3)
		assert.Equal(t, "session-1", sessions[0].Name)
		assert.Equal(t, "session-2", sessions[1].Name)
		assert.Equal(t, "session-3", sessions[2].Name)
	})

	t.Run("parses session with metadata", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				// Zellij output format: "session-name [Created 2h ago] (current)"
				return &exec.Result{
					Stdout:   []byte("my-session [Created 2h ago] (current)\nother-session [Created 1d ago]\n"),
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		sessions, err := z.ListSessions(ctx)

		require.NoError(t, err)
		require.Len(t, sessions, 2)
		assert.Equal(t, "my-session", sessions[0].Name)
		assert.Equal(t, "other-session", sessions[1].Name)
	})

	t.Run("handles no active sessions error gracefully", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stderr:   []byte("No active zellij sessions found"),
					ExitCode: 0, // zellij may return 0 with this message
				}, nil
			},
		}

		z := NewZellij(mockExec)
		sessions, err := z.ListSessions(ctx)

		require.NoError(t, err)
		assert.Empty(t, sessions)
	})

	t.Run("returns error on command failure", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stderr:   []byte("zellij: command not found"),
					ExitCode: 127,
				}, errors.New("exit code 127")
			},
		}

		z := NewZellij(mockExec)
		_, err := z.ListSessions(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "list sessions")
	})
}

func TestZellij_KillSession(t *testing.T) {
	ctx := context.Background()

	t.Run("kills session successfully", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				assert.Equal(t, "zellij", opts.Name)
				assert.Equal(t, []string{"kill-session", "my-session"}, opts.Args)

				return &exec.Result{
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		err := z.KillSession(ctx, "my-session")

		require.NoError(t, err)
	})

	t.Run("returns ErrSessionNotFound when session missing", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stderr:   []byte("Session 'missing' not found"),
					ExitCode: 1,
				}, errors.New("exit code 1")
			},
		}

		z := NewZellij(mockExec)
		err := z.KillSession(ctx, "missing")

		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("returns ErrSessionNotFound for no session error", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stderr:   []byte("No session with that name"),
					ExitCode: 1,
				}, errors.New("exit code 1")
			},
		}

		z := NewZellij(mockExec)
		err := z.KillSession(ctx, "missing")

		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("returns generic error for other failures", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stderr:   []byte("unexpected error"),
					ExitCode: 1,
				}, errors.New("exit code 1")
			},
		}

		z := NewZellij(mockExec)
		err := z.KillSession(ctx, "my-session")

		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrSessionNotFound)
		assert.Contains(t, err.Error(), "kill session")
	})
}

func TestZellij_AttachSession(t *testing.T) {
	ctx := context.Background()

	t.Run("attaches to session with correct args", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				assert.Equal(t, "zellij", opts.Name)
				assert.Equal(t, []string{"attach", "my-session", "--create"}, opts.Args)

				return &exec.Result{
					ExitCode: 0,
				}, nil
			},
		}

		z := NewZellij(mockExec)
		// Note: This test won't fully exercise TTY handling since we're not in a terminal
		err := z.AttachSession(ctx, "my-session")

		require.NoError(t, err)
	})

	t.Run("returns ErrAttachFailed on command error", func(t *testing.T) {
		mockExec := &mocks.ExecutorMock{
			RunFunc: func(ctx context.Context, opts exec.RunOptions) (*exec.Result, error) {
				return &exec.Result{
					Stderr:   []byte("attach failed"),
					ExitCode: 1,
				}, errors.New("exit code 1")
			},
		}

		z := NewZellij(mockExec)
		err := z.AttachSession(ctx, "my-session")

		assert.ErrorIs(t, err, ErrAttachFailed)
	})
}
