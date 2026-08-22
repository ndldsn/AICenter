//go:build windows

package service

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// setProcessGroup runs the command in its own process group on Windows.
func setProcessGroup(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

// killProcessTree terminates a process and all its descendants. On Windows
// os.Process.Kill only reaps the direct child (the shell), so children such as
// `sleep` would survive and keep the stdout pipe open. We therefore use
// `taskkill /T` which walks the whole tree.
func killProcessTree(p *os.Process) error {
	if p == nil {
		return nil
	}
	if err := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(p.Pid)).Run(); err != nil {
		// Fallback to a direct kill if taskkill is unavailable.
		_ = p.Kill()
		return err
	}
	return nil
}
