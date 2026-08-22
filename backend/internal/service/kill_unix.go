//go:build !windows

package service

import (
	"os"
	"os/exec"
	"syscall"
)

// killProcessTree kills the process and its whole process group (negative PID).
func killProcessTree(p *os.Process) error {
	if p == nil {
		return nil
	}
	pgid := getPgid(p)
	// Kill the foreground process first.
	_ = p.Kill()
	if pgid > 0 {
		// Kill the entire process group to reap children (e.g. `sleep`).
		syscall.Kill(-pgid, syscall.SIGKILL)
	}
	return nil
}

func getPgid(p *os.Process) int {
	if p == nil {
		return 0
	}
	pgid, err := syscall.Getsid(p.Pid)
	if err != nil {
		return 0
	}
	return pgid
}

// setProcessGroup runs the command in its own process group so descendants can
// be reaped via kill(-pgid).
func setProcessGroup(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
