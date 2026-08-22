//go:build windows

package terminal

import (
	"io"
	"os"
	"os/exec"
)

func defaultShell() string {
	for _, c := range []string{"bash", "sh"} {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	if c := os.Getenv("COMSPEC"); c != "" {
		return c
	}
	return "cmd"
}

func shellArgs(cmd string) []string {
	if cmd == "bash" || cmd == "sh" {
		return []string{"-i"}
	}
	return nil
}

// startSession launches the shell on Windows via stdin/stdout pipes.
func startSession(sess *Session, shell string, cols, rows int) error {
	c := exec.Command(shell, shellArgs(shell)...)
	stdinPipe, err := c.StdinPipe()
	if err != nil {
		return err
	}
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		stdinPipe.Close()
		return err
	}
	seR, seW, err := os.Pipe()
	if err != nil {
		stdinPipe.Close()
		stdoutPipe.Close()
		return err
	}
	c.Stderr = seW
	if err := c.Start(); err != nil {
		stdinPipe.Close()
		stdoutPipe.Close()
		seR.Close()
		seW.Close()
		return err
	}
	// Stderr writer is owned by cmd and closed when it exits. We only read seR.
	sess.cmd = c
	sess.stdin = stdinPipe
	sess.stdout = &pipeMerger{
		r1: stdoutPipe,
		r2: seR,
		w2: seW,
	}
	return nil
}

// resizePTY is a no-op on the Windows pipe path.
func resizePTY(sess *Session, cols, rows int) error { return nil }

// cleanupPTY is a no-op on the Windows pipe path.
func cleanupPTY(sess *Session) {}

// pipeMerger merges two readers into one; w2 is the *os.Pipe writer that feeds
// r2 (kept so the consumer can close stderr side cleanly).
type pipeMerger struct {
	r1, r2 io.ReadCloser
	w2     *os.File
	done1  bool
	done2  bool
}

func (pm *pipeMerger) Read(p []byte) (int, error) {
	for {
		if !pm.done1 {
			n, err := pm.r1.Read(p)
			if n > 0 {
				return n, nil
			}
			if err != nil {
				pm.done1 = true
			}
		}
		if !pm.done2 {
			n, err := pm.r2.Read(p)
			if n > 0 {
				return n, nil
			}
			if err != nil {
				pm.done2 = true
			}
		}
		if pm.done1 && pm.done2 {
			return 0, io.EOF
		}
	}
}

func (pm *pipeMerger) Close() error {
	_ = pm.r1.Close()
	_ = pm.r2.Close()
	if pm.w2 != nil {
		_ = pm.w2.Close()
	}
	return nil
}

var _ = io.EOF
