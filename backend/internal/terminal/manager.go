// Package terminal manages shell sessions streamed over WebSocket.
//
// On POSIX it uses a real PTY (creack/pty). On Windows it falls back to a
// plain os/exec pipeline (stdin/stdout pipes), which xterm.js drives as a
// non-TTY shell — sufficient for the control-center "run a command in a tab"
// use-case. The WebSocket message envelope is identical on both platforms.
package terminal

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Session is a running shell bridged to the caller over a pipe.
type Session struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"server_id"`
	Command   string    `json:"command"`
	CreatedAt time.Time `json:"created_at"`
	mu        sync.Mutex
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	done      chan struct{}
	exitCode  int
}

func NewManager(log *zap.Logger) *Manager {
	return &Manager{log: log, sessions: map[string]*Session{}}
}

type Manager struct {
	log      *zap.Logger
	mu       sync.RWMutex
	sessions map[string]*Session
}

// Create spawns a shell. On POSIX it allocates a real PTY; on Windows it uses
// stdin/stdout pipes. cols/rows are honored on POSIX (no-op on the pipe path).
func (m *Manager) Create(serverID, shell string, cols, rows int) (*Session, error) {
	if shell == "" {
		shell = defaultShell()
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	sess := &Session{
		ID:        uuid.New().String(),
		ServerID:  serverID,
		Command:   shell,
		CreatedAt: time.Now().UTC(),
		done:      make(chan struct{}),
	}

	var err error
	err = startSession(sess, shell, cols, rows)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = sess.cmd.Wait()
		sess.mu.Lock()
		sess.exitCode = exitCode(sess.cmd)
		sess.mu.Unlock()
		close(sess.done)
	}()

	m.mu.Lock()
	m.sessions[sess.ID] = sess
	m.mu.Unlock()
	return sess, nil
}

func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

func (m *Manager) List() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]map[string]interface{}, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, map[string]interface{}{
			"id":         s.ID,
			"server_id":  s.ServerID,
			"command":    s.Command,
			"created_at": s.CreatedAt,
		})
	}
	return out
}

// Resize changes the window size (POSIX PTY only; no-op elsewhere).
func (s *Session) Resize(cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return resizePTY(s, cols, rows)
}

// Write sends input bytes to the shell's stdin.
func (s *Session) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stdin == nil {
		return 0, os.ErrClosed
	}
	return s.stdin.Write(b)
}

// ReadLoop pumps shell output into sink until the process exits.
func (s *Session) ReadLoop(sink func([]byte)) {
	buf := make([]byte, 4096)
	for {
		select {
		case <-s.done:
			return
		default:
		}
		s.mu.Lock()
		rc := s.stdout
		s.mu.Unlock()
		if rc == nil {
			return
		}
		n, err := rc.Read(buf)
		if n > 0 {
			cp := make([]byte, n)
			copy(cp, buf[:n])
			sink(cp)
		}
		if err != nil {
			return
		}
	}
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	if s.stdout != nil {
		_ = s.stdout.Close()
	}
	if s.stdin != nil || s.stdout != nil {
		cleanupPTY(s)
	}
}

func (m *Manager) remove(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}

func (m *Manager) Remove(id string) {
	if s, ok := m.Get(id); ok {
		s.Close()
		m.remove(id)
	}
}

// ---- message envelope ----

type Message struct {
	Type string `json:"type"` // data | input | resize | exit | error
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
	Code int    `json:"code,omitempty"`
}

// MarshalMessage returns the JSON envelope (used to send typed frames).
func MarshalMessage(m Message) ([]byte, error) {
	return json.Marshal(m)
}

// ExitCode returns the process exit code once done.
func (s *Session) ExitCode() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitCode
}

func exitCode(cmd *exec.Cmd) int {
	if cmd == nil || cmd.ProcessState == nil {
		return 0
	}
	return cmd.ProcessState.ExitCode()
}
