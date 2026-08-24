//go:build !windows

package service

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// killProcessTree kills the process and its whole process group (negative PID).
func killProcessTree(p *os.Process) error {
	if p == nil {
		return nil
	}
	pgid := getPgid(p)
	_ = p.Kill()
	if pgid > 0 {
		_ = unix.Kill(-pgid, unix.SIGKILL)
	}
	return nil
}

func getPgid(p *os.Process) int {
	if p == nil {
		return 0
	}
	pgid, err := unix.Getsid(p.Pid)
	if err != nil {
		return 0
	}
	return pgid
}

func setProcessGroup(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
