package middleware

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aicenter/aicenter/internal/repository"
	"github.com/gin-gonic/gin"
)

// sanitizeKeys is the set of JSON field names that MUST be redacted before
// a request or response body is written into audit_logs. AGENTS.md requires
// audit never persist Password, API Key, Private Key, or Token.
var sanitizeKeys = []string{
	"password",
	"api_key",
	"apikey",
	"private_key",
	"token",
	"refresh_token",
	"authorization",
	"secret",
}

// AuditMiddleware records every authenticated REST request into the audit
// table. Best-effort: audit failure never rejects the user request.
//
// Action classification is based on Method + Path only (not handler name) so
// it survives refactors. Identity (userID/username/role) is read from gin
// context injected by JWTAuth.
func AuditMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var action string
		var resourceType, resourceID, resourceName string

		if permName, ok := c.Get("requiredPermission"); ok {
			action = permName.(string)
			resourceType = strings.Split(action, ".")[0]
			resourceName = action
		} else {
			action = classifyAction(c.Request.Method, c.Request.URL.Path, c.Request.URL.Query())
			resourceType = action
			resourceName = action
		}
		resourceID, resourceType = resolveResourceID(c)

		// Wrap the response writer so we can capture the response body for
		// audit. The wrapped writer forwards everything to the real writer.
		aw := &auditWriter{ResponseWriter: c.Writer}
		c.Writer = aw

		c.Next()

		statusCode := c.Writer.Status()
		durationMs := int(time.Since(start).Milliseconds())

		var userID, username string
		if u, ok := c.Get("userID"); ok {
			userID = u.(string)
		}
		if u, ok := c.Get("username"); ok {
			username = u.(string)
		}
		// role is intentionally not written: audit_logs table has no role column

		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			// Restore so downstream handlers (ShouldBindJSON etc.) still work.
			c.Request.Body = io.NopCloser(bytes.NewReader(reqBody))
		}
		sanitizedReq := sanitizeBody(reqBody)

		var respBody []byte
		if w, ok := c.Writer.(*auditWriter); ok {
			respBody = sanitizeBody(w.buf)
		}

		errMsg := ""
		if errs := c.Errors; errs != nil && len(errs) > 0 {
			errMsg = errs.Last().Error()
			if len(errMsg) > 1024 {
				errMsg = errMsg[:1024]
			}
		}

		go func() {
			time.Sleep(100 * time.Millisecond)
			repo := repository.NewAuditRepository(db)
			entry := repository.AuditEntry{
				UserID:       userID,
				Username:     username,
				Action:       action,
				ResourceType: resourceType,
				ResourceID:   resourceID,
				ResourceName: resourceName,
				Method:       c.Request.Method,
				Path:         c.Request.URL.Path,
				IP:           c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				StatusCode:   statusCode,
				RequestBody:  strings.TrimSpace(string(sanitizedReq)),
				ResponseBody: strings.TrimSpace(string(respBody)),
				Error:        errMsg,
				DurationMs:   durationMs,
			}
			_ = repo.Record(&entry)
		}()
	}
}

// sanitizeBody redacts known-secret keys from a JSON body before it is
// written to audit_logs. AGENTS.md forbids recording Password/API Key/
// Private Key/Token.
func sanitizeBody(body []byte) []byte {
	if len(body) == 0 {
		return body
	}
	s := string(body)
	for _, key := range sanitizeKeys {
		s = redactJSONString(s, key)
	}
	return []byte(s)
}

// redactJSONString replaces all occurrences of the JSON key's value with
// "[REDACTED]". It advances through the string so it never re-visits an
// already-redacted segment (avoiding infinite loops on repeated keys).
func redactJSONString(s, key string) string {
	prefix := `"` + key + `":`
	var out strings.Builder
	for len(s) > 0 {
		idx := strings.Index(s, prefix)
		if idx < 0 {
			return out.String() + s
		}
		out.WriteString(s[:idx+len(prefix)])
		s = s[idx+len(prefix):]
		// skip whitespace
		for len(s) > 0 && (s[0] == ' ' || s[0] == '	' || s[0] == '\n') {
			out.WriteByte(s[0])
			s = s[1:]
		}
		if len(s) == 0 {
			return out.String()
		}
		switch s[0] {
		case '"':
			end := strings.Index(s[1:], `"`)
			if end < 0 {
				return out.String() + s
			}
			out.WriteString(`"[REDACTED]"`)
			s = s[end+2:]
		case '{', '[':
			var close byte
			if s[0] == '{' {
				close = '}'
			} else {
				close = ']'
			}
			end := findClosing(s, s[0], close)
			if end < 0 {
				return out.String() + s
			}
			out.WriteString(`"[REDACTED]"`)
			s = s[end+1:]
		case 'n':
			if strings.HasPrefix(s, "null") {
				out.WriteString(`"[REDACTED]"`)
				s = s[4:]
			} else {
				return out.String() + s
			}
		case 't':
			if strings.HasPrefix(s, "true") {
				out.WriteString(`"[REDACTED]"`)
				s = s[4:]
			} else {
				return out.String() + s
			}
		case 'f':
			if strings.HasPrefix(s, "false") {
				out.WriteString(`"[REDACTED]"`)
				s = s[5:]
			} else {
				return out.String() + s
			}
		default:
			end := 0
			for end < len(s) && s[end] != ',' && s[end] != '}' && s[end] != ']' {
				end++
			}
			out.WriteString(`"[REDACTED]"`)
			s = s[end:]
		}
	}
	return out.String()
}

func findClosing(s string, open, close byte) int {
	nest := 1
	inStr := false
	for i := 1; i < len(s); i++ {
		ch := s[i]
		if inStr {
			if ch == '\\' {
				i++ // skip escaped
				continue
			}
			inStr = ch != '"'
			continue
		}
		switch ch {
		case '"':
			inStr = true
		case open:
			nest++
		case close:
			nest--
			if nest == 0 {
				return i
			}
		}
	}
	return -1
}

// classifyAction maps HTTP method + path to a stable audit action string.
func classifyAction(method, path string, q url.Values) string {
	m := strings.ToUpper(method)
	raw := strings.TrimRight(path, "/")
	if idx := strings.Index(raw, "?"); idx >= 0 {
		raw = raw[:idx]
	}
	p := raw

	switch p {
	case "/auth/login":
		return "auth.login"
	case "/auth/register":
		return "auth.register"
	case "/auth/refresh":
		return "auth.refresh"
	case "/auth/me":
		return "auth.me"
	case "/dashboard":
		return "dashboard.get"
	}

	if strings.HasPrefix(p, "/settings") {
		if m == "PUT" {
			return "settings.update"
		}
		return "settings.get"
	}

	if strings.HasPrefix(p, "/users") {
		if m == "PATCH" && strings.HasSuffix(p, "/role") {
			return "users.role"
		}
		return resourceVerb("users", m, hasID(p, "/users"))
	}

	if strings.HasPrefix(p, "/roles") {
		if strings.HasSuffix(p, "/permissions") {
			return "roles_permissions.assign"
		}
		if strings.HasSuffix(p, "/groups") {
			return "roles.groups"
		}
		return resourceVerb("roles", m, hasID(p, "/roles"))
	}
	if strings.HasPrefix(p, "/permissions") {
		return "permissions.list"
	}

	if strings.HasPrefix(p, "/servers") {
		if strings.HasSuffix(p, "/restart") {
			return "servers.restart"
		}
		if strings.HasSuffix(p, "/ping") {
			return "servers.ping"
		}
		return resourceVerb("servers", m, hasID(p, "/servers"))
	}

	if strings.HasPrefix(p, "/docker/containers") {
		if strings.Contains(p, "/start/") {
			return "docker_containers.start"
		}
		if strings.Contains(p, "/stop/") {
			return "docker_containers.stop"
		}
		if strings.Contains(p, "/restart/") {
			return "docker_containers.restart"
		}
		if strings.HasSuffix(p, "/prune") {
			return "docker_containers.prune"
		}
		if strings.HasSuffix(p, "/run") {
			return "docker_containers.run"
		}
		if strings.HasSuffix(p, "/exec") {
			return "docker_containers.exec"
		}
		if strings.HasSuffix(p, "/logs") {
			return "docker_containers.logs"
		}
		return resourceVerb("docker_containers", m, hasID(p, "/docker/containers"))
	}

	if strings.HasPrefix(p, "/docker/images") {
		if q.Get("pull") == "true" {
			return "docker_images.pull"
		}
		if strings.HasSuffix(p, "/prune") {
			return "docker_images.prune"
		}
		if strings.HasSuffix(p, "/build") {
			return "docker_images.build"
		}
		return resourceVerb("docker_images", m, hasID(p, "/docker/images"))
	}

	if strings.HasPrefix(p, "/docker/volumes") {
		if strings.HasPrefix(p, "/docker/volumes/prune") {
			return "docker_volumes.prune"
		}
		if m == "POST" && !hasID(p, "/docker/volumes") {
			return "docker_volumes.create"
		}
		return resourceVerb("docker_volumes", m, hasID(p, "/docker/volumes"))
	}

	if strings.HasPrefix(p, "/ai/providers") {
		return resourceVerb("ai_providers", m, hasID(p, "/ai/providers"))
	}
	if strings.HasPrefix(p, "/ai/models") {
		return "ai_models.list"
	}
	if p == "/ai/chat" {
		return "ai_chat.send"
	}

	if strings.HasPrefix(p, "/agents/sessions") {
		if strings.HasSuffix(p, "/messages") {
			return "agents_sessions.send"
		}
		return resourceVerb("agents_sessions", m, hasID(p, "/agents/sessions"))
	}
	if strings.HasPrefix(p, "/agents") {
		if strings.HasSuffix(p, "/sessions") {
			return "agents_sessions.create"
		}
		return resourceVerb("agents", m, hasID(p, "/agents"))
	}

	if strings.HasPrefix(p, "/audit-logs") {
		return "audit.list"
	}
	if strings.HasPrefix(p, "/approvals") {
		if strings.HasSuffix(p, "/approve") {
			return "approvals.approve"
		}
		if strings.HasSuffix(p, "/reject") {
			return "approvals.reject"
		}
		return resourceVerb("approvals", m, hasID(p, "/approvals"))
	}

	if strings.HasPrefix(p, "/tasks") {
		if strings.HasSuffix(p, "/retry") {
			return "tasks.retry"
		}
		if strings.HasSuffix(p, "/cancel") {
			return "tasks.cancel"
		}
		return resourceVerb("tasks", m, hasID(p, "/tasks"))
	}

	if strings.HasPrefix(p, "/terminal/sessions") {
		if strings.HasSuffix(p, "/close") {
			return "terminal.close"
		}
		if m == "POST" && !hasID(p, "/terminal/sessions") {
			return "terminal.create"
		}
		return resourceVerb("terminal", m, hasID(p, "/terminal/sessions"))
	}

	if strings.HasPrefix(p, "/notifications") {
		return resourceVerb("notifications", m, hasID(p, "/notifications"))
	}
	if strings.HasPrefix(p, "/monitor") {
		return resourceVerb("monitor", m, hasID(p, "/monitor"))
	}

	return fmt.Sprintf("%s.%s", strings.ToLower(m), strings.ReplaceAll(p, "/", "_"))
}

func resolveResourceID(c *gin.Context) (string, string) {
	if idParam, ok := c.Params.Get("id"); ok {
		return idParam, ""
	}
	params := c.Params
	if len(params) == 0 {
		return "", ""
	}
	last := params[len(params)-1]
	if last.Key == "id" {
		return last.Value, ""
	}
	return last.Value, ""
}

func resourceVerb(resource, method string, hasID bool) string {
	switch {
	case method == "GET" && !hasID:
		return resource + ".list"
	case method == "GET" && hasID:
		return resource + ".get"
	case method == "POST" && !hasID:
		return resource + ".create"
	case method == "PUT" || method == "PATCH":
		return resource + ".update"
	case method == "DELETE":
		return resource + ".delete"
	default:
		return resource + "." + strings.ToLower(method)
	}
}

func hasID(path, base string) bool {
	return strings.HasPrefix(strings.TrimRight(path, "/"), base+"/")
}

// auditWriter captures response bytes for audit.
type auditWriter struct {
	gin.ResponseWriter
	buf []byte
}

func (w *auditWriter) Write(b []byte) (int, error) {
	w.buf = append(w.buf, b...)
	return w.ResponseWriter.Write(b)
}