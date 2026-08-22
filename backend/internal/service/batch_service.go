package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/pkg/ssh"
	"github.com/aicenter/aicenter/internal/repository"
)

// BatchService runs a single command across many servers in parallel and
// collects per-server results.
type BatchService struct {
	repo *repository.ServerRepository
}

func NewBatchService() *BatchService {
	return &BatchService{repo: repository.NewServerRepository()}
}

// BatchResult is the outcome of running `command` on a single server.
type BatchResult struct {
	ServerID  string `json:"server_id"`
	Server    string `json:"server"`
	Host      string `json:"host"`
	Status    string `json:"status"`    // "ok" | "failed"
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	Error     string `json:"error,omitempty"`
	Duration  string `json:"duration"` // e.g. "1.234s"
	ExitCode  int    `json:"exit_code"`
}

// BatchRequest is the payload for POST /servers/batch/command.
type BatchRequest struct {
	Command  string   `json:"command"`  // required
	ServerIDs []string `json:"server_ids"` // nil/empty = all servers
	Timeout  int      `json:"timeout_seconds"` // 0 = 30s default
	GroupID  *string  `json:"group_id,omitempty"`
}

// BatchCommand runs the command across servers concurrently and returns the
// aggregated per-server results (ordered as requested).
func (s *BatchService) BatchCommand(ctx context.Context, req *BatchRequest) []*BatchResult {
	if req.Command == "" {
		return nil
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	servers, _, err := s.repo.List(0, 10000)
	if err != nil {
		return []*BatchResult{{Status: "failed", Error: "list servers: " + err.Error()}}
	}
	if len(req.ServerIDs) > 0 && req.GroupID != nil && *req.GroupID != "" {
		var filtered []*models.Server
		want := map[string]bool{}
		for _, id := range req.ServerIDs {
			want[id] = true
		}
		for _, sv := range servers {
			if want[sv.ID] {
				filtered = append(filtered, sv)
			}
		}
		servers = filtered
	} else if len(req.ServerIDs) > 0 {
		want := map[string]bool{}
		for _, id := range req.ServerIDs {
			want[id] = true
		}
		ordered := make([]*models.Server, 0, len(req.ServerIDs))
		for _, id := range req.ServerIDs {
			for _, sv := range servers {
				if sv.ID == id {
					ordered = append(ordered, sv)
					break
				}
			}
		}
		servers = ordered
	}

	results := make([]*BatchResult, len(servers))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i, sv := range servers {
		wg.Add(1)
		go func(i int, sv *models.Server) {
			defer wg.Done()
			res := s.runOn(ctx, sv, req.Command, timeout)
			mu.Lock()
			results[i] = res
			mu.Unlock()
		}(i, sv)
	}
	wg.Wait()
	return results
}

// runOn executes the command on a single server, via local exec for
// localhost/127.0.0.1 (verifiable without an SSH daemon) or SSH otherwise.
func (s *BatchService) runOn(parent context.Context, sv *models.Server, command string, timeout int) *BatchResult {
	res := &BatchResult{
		ServerID: sv.ID,
		Server:   sv.Name,
		Host:     fmt.Sprintf("%s:%d", sv.Host, sv.Port),
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(parent, time.Duration(timeout)*time.Second)
	defer cancel()

	stdout, stderr, rc, err := runShell(ctx, sv, command)
	res.Duration = time.Since(start).Round(time.Millisecond).String()
	res.Stdout = stdout
	res.Stderr = stderr
	if rc != nil {
		res.ExitCode = *rc
	}
	if err != nil {
		res.Status = "failed"
		res.Error = err.Error()
		if ctx.Err() == context.DeadlineExceeded {
			res.Error = fmt.Sprintf("timeout after %ds", timeout)
		}
		return res
	}
	res.Status = "ok"
	return res
}

// runShell dispatches between local exec (localhost) and SSH (remote).
func runShell(ctx context.Context, sv *models.Server, command string) (stdout, stderr string, rc *int, err error) {
	h := strings.ToLower(strings.TrimSpace(sv.Host))
	switch h {
	case "", "localhost", "127.0.0.1", "::1":
		return runLocal(ctx, command)
	default:
		return runSSH(ctx, sv, command)
	}
}

// runLocal executes the command on the backend host. The timeout is enforced
// via a deadline watcher that kills the whole process tree (including children
// spawned by the shell, e.g. `sleep`) so CombinedOutput returns promptly.
func runLocal(ctx context.Context, command string) (string, string, *int, error) {
	c := exec.Command("sh", "-c", command)
	setProcessGroup(c)
	var outBuf bytes.Buffer
	c.Stdout = &outBuf
	c.Stderr = &outBuf
	if err := c.Start(); err != nil {
		return "", "", nil, err
	}

	// Watcher: if the context expires first, kill the process AND its tree.
	aborted, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done():
			killProcessTree(c.Process)
		case <-aborted.Done():
		}
	}()
	defer cancel()

	err := c.Wait()
	if e, ok := err.(*exec.ExitError); ok {
		// Did we hit the deadline and force-kill?
		if ctx.Err() != nil {
			return outBuf.String(), "", nil, fmt.Errorf("timeout: %w", ctx.Err())
		}
		rc := e.ExitCode()
		return outBuf.String(), "", &rc, nil
	}
	return outBuf.String(), "", nil, nil
}

// runSSH dials the server over SSH using the stored credentials and runs the
// command. PasswordEnc / PrivateKeyEnc are passed through (the model still
// TODOs encryption, matching the existing TestConnection path).
func runSSH(ctx context.Context, sv *models.Server, command string) (string, string, *int, error) {
	cfg := &ssh.Config{
		Host:       sv.Host,
		Port:       sv.Port,
		Username:   sv.Username,
		AuthType:   sv.AuthType,
		Password:   sv.PasswordEnc,
		PrivateKey: sv.PrivateKeyEnc,
		Timeout:    10 * time.Second,
	}
	client := ssh.NewClient(cfg)
	if err := client.Connect(); err != nil {
		return "", "", nil, fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()
	// SSH session has no context-awareness; respect cancellation via a final
	// close rather than a hard kill.
	stdout, stderr, err := client.Run(command)
	rc := 0
	if err != nil {
		rc = 1
	}
	return stdout, stderr, &rc, err
}

// sshRun is a small helper kept for callers that already hold an ssh.Config.
func sshRun(cfg *ssh.Config, command string) (string, string, error) {
	client := ssh.NewClient(cfg)
	if err := client.Connect(); err != nil {
		return "", "", fmt.Errorf("ssh connect: %w", err)
	}
	defer client.Close()
	return client.Run(command)
}
