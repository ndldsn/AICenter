//go:build !windows

package terminal

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func defaultShell() string {
	for _, c := range []string{"/bin/bash", "/bin/zsh", "/bin/sh"} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "sh"
}

func shellArgs(cmd string) []string {
	return []string{"-i"}
}

// startSession launches the shell under a pseudo-TTY (POSIX).
func startSession(sess *Session, shell string, cols, rows int) error {
	c := exec.Command(shell, shellArgs(shell)...)
	tty, err := pty.StartWithSize(c, &pty.Winsize{
		Cols: uint16(cols), Rows: uint16(rows),
	})
	if err != nil {
		return err
	}
	sess.cmd = c
	sess.stdin = tty
	sess.stdout = tty
	return nil
}

// resizePTY resizes the PTY window (POSIX only).
func resizePTY(sess *Session, cols, rows int) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.stdin == nil {
		return nil
	}
	if f, ok := sess.stdin.(*os.File); ok {
		return pty.Setsize(f, &pty.Winsize{
			Cols: uint16(cols), Rows: uint16(rows),
		})
	}
	return nil
}

// cleanupPTY is a no-op for the unix PTY path (stdin/stdout already closed).
func cleanupPTY(sess *Session) {}
