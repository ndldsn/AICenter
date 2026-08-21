package ssh

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds SSH connection configuration
type Config struct {
	Host       string
	Port       int
	Username   string
	AuthType   string // "password" or "key"
	Password   string
	PrivateKey string
	Timeout    time.Duration
}

// Client represents an SSH connection wrapper
type Client struct {
	config *Config
	client *ssh.Client
}

// NewClient creates a new SSH client
func NewClient(cfg *Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Client{config: cfg}
}

// Connect establishes the SSH connection
func (c *Client) Connect() error {
	auth, err := c.getAuthMethod()
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            c.config.Username,
		Auth:            []ssh.AuthMethod{auth},
		Timeout:         c.config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Use known_hosts in production
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	c.client = client
	return nil
}

// Run executes a command on the remote server
func (c *Client) Run(command string) (string, string, error) {
	if c.client == nil {
		return "", "", fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return stdout.String(), stderr.String(), err
	}

	return stdout.String(), stderr.String(), nil
}

// Close closes the connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *Client) getAuthMethod() (ssh.AuthMethod, error) {
	switch c.config.AuthType {
	case "key":
		if c.config.PrivateKey != "" {
			signer, err := ssh.ParsePrivateKey([]byte(c.config.PrivateKey))
			if err != nil {
				// Try with passphrase
				if _, ok := err.(*ssh.PassphraseMissingError); ok {
					signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(c.config.PrivateKey), []byte(""))
				}
				if err != nil {
					return nil, fmt.Errorf("parse key: %w", err)
				}
			}
			return ssh.PublicKeys(signer), nil
		}

		// Try default key files
		home, _ := os.UserHomeDir()
		keyFiles := []string{
			fmt.Sprintf("%s/.ssh/id_rsa", home),
			fmt.Sprintf("%s/.ssh/id_ed25519", home),
		}
		for _, keyFile := range keyFiles {
			if _, err := os.Stat(keyFile); err == nil {
				key, err := os.ReadFile(keyFile)
				if err != nil {
					continue
				}
				signer, err := ssh.ParsePrivateKey(key)
				if err != nil {
					continue
				}
				return ssh.PublicKeys(signer), nil
			}
		}

		return nil, fmt.Errorf("no private key found")

	default: // password
		return ssh.Password(c.config.Password), nil
	}
}

// TestConnection tests an SSH connection without keeping it open
func TestConnection(cfg *Config) error {
	client := NewClient(cfg)
	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	_, _, err := client.Run("echo 'SSH connection successful'")
	return err
}

// ParsePrivateKey parses and validates a private key
func ParsePrivateKey(keyPEM string) (bool, error) {
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		// Try parsing without PEM header
		_, err := ssh.ParsePrivateKey([]byte(keyPEM))
		if err != nil {
			return false, fmt.Errorf("invalid private key format")
		}
		return true, nil
	}

	// Check for encrypted key
	if x509.IsEncryptedPEMBlock(block) {
		return false, fmt.Errorf("encrypted private key requires passphrase")
	}

	_, err := ssh.ParsePrivateKey([]byte(keyPEM))
	if err != nil {
		return false, fmt.Errorf("invalid private key: %w", err)
	}

	return true, nil
}

// GetSSHVersion attempts to get the SSH server banner
func GetSSHVersion(host string, port int) (string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Read the banner
	buf := make([]byte, 256)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := io.ReadAtLeast(conn, buf, 4)
	if err != nil {
		return "", err
	}

	banner := strings.TrimSpace(string(buf[:n]))
	if !strings.HasPrefix(banner, "SSH-") {
		return "", fmt.Errorf("not an SSH server: %s", banner)
	}

	return banner, nil
}
