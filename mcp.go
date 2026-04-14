// Package protocol — MCP server management wire types.
//
// This file defines the wire protocol for managing MCP (Model Context
// Protocol) servers from the PWA via the AgentD daemon. Added in
// feature 169.
//
// Both agentd and agentd-relay import these types via type aliases in
// their respective internal/relay/types.go files. The TypeScript mirror
// lives in agentd-web/src/types/index.ts and must be kept in lockstep
// with the Go JSON tags.

package protocol

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// ─── Enum: MCPServerScope ─────────────────────────────────────────────────

// MCPServerScope identifies where an MCP server configuration is persisted.
// `project` entries live in `<workdir>/.mcp.json`; `user` entries live in
// `~/.claude.json`. Claude Code CLI's third scope (`local`) is explicitly
// out of scope for feature 169 — see spec.md Non-Goals.
type MCPServerScope string

// MCPServerScope values.
const (
	// MCPScopeProject persists the server in <workdir>/.mcp.json.
	MCPScopeProject MCPServerScope = "project"
	// MCPScopeUser persists the server in ~/.claude.json.
	MCPScopeUser MCPServerScope = "user"
)

// ─── Wire Types ───────────────────────────────────────────────────────────

// MCPServerConfig is the canonical representation of a single MCP server
// configuration. Field set depends on Transport:
//
//   - transport=stdio : Command (required), Args (optional), Env (optional)
//   - transport=sse   : URL (required), Headers (optional)
//   - transport=http  : URL (required), Headers (optional)
//
// Env and Headers values are secrets — never log them.
type MCPServerConfig struct {
	Name      string            `json:"name"`
	Scope     MCPServerScope    `json:"scope"`
	Transport string            `json:"transport"`
	Enabled   bool              `json:"enabled"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	URL       string            `json:"url,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// MCPServerStatusEntry is the live connection state for a single MCP server
// as reported by the Go SDK's client.MCPServerStatus(). Not persisted — the
// daemon computes it on demand. `disabled` is synthesized locally when a
// config entry has Enabled=false (the SDK has no entry for those servers).
type MCPServerStatusEntry struct {
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	Message    string         `json:"message,omitempty"`
	ToolCount  int            `json:"tool_count,omitempty"`
	ShadowedBy MCPServerScope `json:"shadowed_by,omitempty"`
}

// MCPListPayload is the PWA → daemon request for the merged MCP server list.
// Empty Scope means both scopes; SessionID is required for project-scope
// resolution.
type MCPListPayload struct {
	Scope     MCPServerScope `json:"scope,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
}

// MCPListResponse is the daemon → PWA response to mcp_list_servers. Servers
// is ordered deterministically: user scope first, then project, sorted by
// name within each scope. Status is empty when no session is active.
type MCPListResponse struct {
	Servers        []MCPServerConfig      `json:"servers"`
	Status         []MCPServerStatusEntry `json:"status,omitempty"`
	ProjectWorkdir string                 `json:"project_workdir,omitempty"`
}

// MCPMutationPayload is the PWA → daemon request for add/update operations.
// SessionID is required for project-scope mutations so the daemon can
// resolve the target workdir; user-scope mutations may omit it.
type MCPMutationPayload struct {
	SessionID string          `json:"session_id,omitempty"`
	Server    MCPServerConfig `json:"server"`
}

// MCPMutationResponse is the daemon → PWA response for add/update/remove/
// toggle/reconnect operations. AppliedToLive is true when the SDK's
// SetMCPServers call succeeded against a live session; false for user-scope
// mutations without an active session, or when the session ended mid-op.
type MCPMutationResponse struct {
	OK            bool                   `json:"ok"`
	Error         string                 `json:"error,omitempty"`
	ErrorCode     string                 `json:"error_code,omitempty"`
	AppliedToLive bool                   `json:"applied_to_live"`
	LiveStatus    []MCPServerStatusEntry `json:"live_status,omitempty"`
	Server        *MCPServerConfig       `json:"server,omitempty"`
}

// MCPRemovePayload is the PWA → daemon request to delete a server from a
// given scope. Shares MCPMutationResponse for the reply.
type MCPRemovePayload struct {
	SessionID string         `json:"session_id,omitempty"`
	Scope     MCPServerScope `json:"scope"`
	Name      string         `json:"name"`
}

// MCPTogglePayload is the PWA → daemon request to enable/disable a server
// without rewriting its full config. Shares MCPMutationResponse for the
// reply.
type MCPTogglePayload struct {
	SessionID string         `json:"session_id,omitempty"`
	Scope     MCPServerScope `json:"scope"`
	Name      string         `json:"name"`
	Enabled   bool           `json:"enabled"`
}

// MCPReconnectPayload is the PWA → daemon request to force a reconnect of
// a named MCP server on a live session. SessionID is required — reconnect
// is session-scoped. Shares MCPMutationResponse for the reply.
type MCPReconnectPayload struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
}

// MCPServersChangedPayload is the daemon → PWA broadcast emitted after every
// successful mutation. Clients re-dispatch mcp_list_servers to refresh their
// view. Action is diagnostic only (add|update|remove|toggle|reconnect).
type MCPServersChangedPayload struct {
	Scope      MCPServerScope `json:"scope"`
	SessionID  string         `json:"session_id,omitempty"`
	Action     string         `json:"action"`
	ServerName string         `json:"server_name"`
}

// ─── Sentinel errors ──────────────────────────────────────────────────────

// Validation sentinel errors. Use errors.Is to check for a specific class;
// the wrapper error from Validate contains additional context.
var (
	// ErrInvalidName indicates a server Name failed the name regex or a
	// reserved-name check.
	ErrInvalidName = errors.New("mcp: invalid server name")
	// ErrInvalidTransport indicates Transport is not stdio/sse/http.
	ErrInvalidTransport = errors.New("mcp: invalid transport")
	// ErrInvalidScope indicates Scope is not project/user.
	ErrInvalidScope = errors.New("mcp: invalid scope")
	// ErrInvalidCommand indicates a stdio Command is empty, too long, or
	// contains shell metacharacters.
	ErrInvalidCommand = errors.New("mcp: invalid command")
	// ErrInvalidArgs indicates stdio Args exceed count or length limits.
	ErrInvalidArgs = errors.New("mcp: invalid args")
	// ErrInvalidEnv indicates stdio Env has too many keys, or a key/value
	// that fails validation.
	ErrInvalidEnv = errors.New("mcp: invalid env")
	// ErrInvalidURL indicates sse/http URL is missing, unparseable, too long,
	// or uses a disallowed scheme.
	ErrInvalidURL = errors.New("mcp: invalid url")
	// ErrInvalidHeaders indicates sse/http Headers has too many keys or a
	// malformed key/value.
	ErrInvalidHeaders = errors.New("mcp: invalid headers")
	// ErrInvalidTransportFields indicates transport-exclusive fields are
	// populated for the wrong transport (e.g., URL set on stdio, Command set
	// on sse).
	ErrInvalidTransportFields = errors.New("mcp: transport-exclusive fields mismatched")
)

// ─── Validation ───────────────────────────────────────────────────────────

// Validation constants used by the rules below.
const (
	mcpMaxNameLen       = 64
	mcpMaxCommandLen    = 2048
	mcpMaxArgs          = 64
	mcpMaxArgLen        = 1024
	mcpMaxEnvKeys       = 32
	mcpMaxEnvValueLen   = 4096
	mcpMaxHeaderKeys    = 32
	mcpMaxHeaderValLen  = 4096
	mcpMaxURLLen        = 2048
)

var (
	mcpNameRe        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
	mcpEnvKeyRe      = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	mcpHeaderTokenRe = regexp.MustCompile("^[A-Za-z0-9!#$%&'*+\\-.^_`|~]+$")

	mcpReservedNames = map[string]struct{}{
		"..":        {},
		".mcp.json": {},
	}

	mcpShellMetacharacters = []string{"|", "&", ";", "$", "`", ">", "<"}

	mcpLoopbackHosts = map[string]struct{}{
		"localhost": {},
		"127.0.0.1": {},
		"::1":       {},
	}
)

// ValidateName checks that name matches the MCP server name rules: byte
// length ≤64, matches ^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$, is not a reserved
// filename, and contains no path separators or leading dot.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("protocol.ValidateName: name is empty: %w", ErrInvalidName)
	}
	if _, reserved := mcpReservedNames[name]; reserved {
		return fmt.Errorf("protocol.ValidateName: %q is reserved: %w", name, ErrInvalidName)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("protocol.ValidateName: %q contains path separator: %w", name, ErrInvalidName)
	}
	if len(name) > mcpMaxNameLen {
		return fmt.Errorf("protocol.ValidateName: %q exceeds %d bytes: %w", name, mcpMaxNameLen, ErrInvalidName)
	}
	if !mcpNameRe.MatchString(name) {
		return fmt.Errorf("protocol.ValidateName: %q does not match name regex: %w", name, ErrInvalidName)
	}
	return nil
}

// ValidateTransport checks that transport is exactly stdio, sse, or http.
// The Claude Code CLI also supports an in-process `sdk` transport, which is
// explicitly out of scope for user-manageable MCP servers — see spec Non-Goals.
func ValidateTransport(transport string) error {
	switch transport {
	case "stdio", "sse", "http":
		return nil
	default:
		return fmt.Errorf("protocol.ValidateTransport: %q is not one of stdio|sse|http: %w", transport, ErrInvalidTransport)
	}
}

// ValidateScope checks that scope is exactly project or user. The CLI's
// third scope (`local`) is not supported by feature 169.
func ValidateScope(scope MCPServerScope) error {
	switch scope {
	case MCPScopeProject, MCPScopeUser:
		return nil
	default:
		return fmt.Errorf("protocol.ValidateScope: %q is not project|user: %w", scope, ErrInvalidScope)
	}
}

// ValidateStdio checks the stdio-transport fields of cfg: Command must be
// non-empty, ≤2048 bytes, and contain no shell metacharacters; Args must
// respect count and length limits; Env keys must match ^[A-Z_][A-Z0-9_]*$
// and values must be ≤4096 bytes; there must be no URL or Headers set.
func ValidateStdio(cfg *MCPServerConfig) error {
	if cfg == nil {
		return fmt.Errorf("protocol.ValidateStdio: cfg is nil: %w", ErrInvalidCommand)
	}
	if cfg.Command == "" {
		return fmt.Errorf("protocol.ValidateStdio: command is empty: %w", ErrInvalidCommand)
	}
	if len(cfg.Command) > mcpMaxCommandLen {
		return fmt.Errorf("protocol.ValidateStdio: command exceeds %d bytes: %w", mcpMaxCommandLen, ErrInvalidCommand)
	}
	for _, meta := range mcpShellMetacharacters {
		if strings.Contains(cfg.Command, meta) {
			return fmt.Errorf("protocol.ValidateStdio: command contains shell metacharacter %q: %w", meta, ErrInvalidCommand)
		}
	}

	if len(cfg.Args) > mcpMaxArgs {
		return fmt.Errorf("protocol.ValidateStdio: args has %d entries, max %d: %w", len(cfg.Args), mcpMaxArgs, ErrInvalidArgs)
	}
	for i, a := range cfg.Args {
		if len(a) > mcpMaxArgLen {
			return fmt.Errorf("protocol.ValidateStdio: args[%d] exceeds %d bytes: %w", i, mcpMaxArgLen, ErrInvalidArgs)
		}
	}

	if len(cfg.Env) > mcpMaxEnvKeys {
		return fmt.Errorf("protocol.ValidateStdio: env has %d keys, max %d: %w", len(cfg.Env), mcpMaxEnvKeys, ErrInvalidEnv)
	}
	for k, v := range cfg.Env {
		if !mcpEnvKeyRe.MatchString(k) {
			return fmt.Errorf("protocol.ValidateStdio: env key %q does not match ^[A-Z_][A-Z0-9_]*$: %w", k, ErrInvalidEnv)
		}
		if len(v) > mcpMaxEnvValueLen {
			// Never log the value itself — it is a secret.
			return fmt.Errorf("protocol.ValidateStdio: env[%s] value exceeds %d bytes: %w", k, mcpMaxEnvValueLen, ErrInvalidEnv)
		}
	}

	if cfg.URL != "" || len(cfg.Headers) > 0 {
		return fmt.Errorf("protocol.ValidateStdio: stdio transport must not set url/headers: %w", ErrInvalidTransportFields)
	}
	return nil
}

// ValidateRemote checks the sse/http-transport fields of cfg: URL must be
// non-empty, ≤2048 bytes, parseable via net/url.Parse, and either use
// scheme=https OR have a loopback host (localhost/127.0.0.1/::1) with any
// scheme. Headers must respect count, RFC 7230 token syntax for keys, and
// length limits for values. Command/Args/Env must be empty.
func ValidateRemote(cfg *MCPServerConfig) error {
	if cfg == nil {
		return fmt.Errorf("protocol.ValidateRemote: cfg is nil: %w", ErrInvalidURL)
	}
	if cfg.URL == "" {
		return fmt.Errorf("protocol.ValidateRemote: url is empty: %w", ErrInvalidURL)
	}
	if len(cfg.URL) > mcpMaxURLLen {
		return fmt.Errorf("protocol.ValidateRemote: url exceeds %d bytes: %w", mcpMaxURLLen, ErrInvalidURL)
	}
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return fmt.Errorf("protocol.ValidateRemote: parsing url: %v: %w", err, ErrInvalidURL)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("protocol.ValidateRemote: url missing scheme or host: %w", ErrInvalidURL)
	}
	host := u.Hostname()
	_, isLoopback := mcpLoopbackHosts[host]
	if !isLoopback {
		// Also treat any IP literal that IsLoopback() as loopback for
		// hosts like ::1 which url.Hostname strips brackets from.
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			isLoopback = true
		}
	}
	if u.Scheme != "https" && !isLoopback {
		return fmt.Errorf("protocol.ValidateRemote: scheme %q requires https or loopback host: %w", u.Scheme, ErrInvalidURL)
	}

	if len(cfg.Headers) > mcpMaxHeaderKeys {
		return fmt.Errorf("protocol.ValidateRemote: headers has %d keys, max %d: %w", len(cfg.Headers), mcpMaxHeaderKeys, ErrInvalidHeaders)
	}
	for k, v := range cfg.Headers {
		if !mcpHeaderTokenRe.MatchString(k) {
			return fmt.Errorf("protocol.ValidateRemote: header key %q is not a valid RFC 7230 token: %w", k, ErrInvalidHeaders)
		}
		if len(v) > mcpMaxHeaderValLen {
			// Never log the value itself — it is a secret.
			return fmt.Errorf("protocol.ValidateRemote: header[%s] value exceeds %d bytes: %w", k, mcpMaxHeaderValLen, ErrInvalidHeaders)
		}
	}

	if cfg.Command != "" || len(cfg.Args) > 0 || len(cfg.Env) > 0 {
		return fmt.Errorf("protocol.ValidateRemote: sse/http transport must not set command/args/env: %w", ErrInvalidTransportFields)
	}
	return nil
}

// Validate runs the full validation suite for an MCPServerConfig and
// returns the first sentinel-wrapped error it finds. Callers should use
// errors.Is to match the specific ErrInvalidXxx class.
func Validate(cfg *MCPServerConfig) error {
	if cfg == nil {
		return fmt.Errorf("protocol.Validate: cfg is nil: %w", ErrInvalidName)
	}
	if err := ValidateName(cfg.Name); err != nil {
		return err
	}
	if err := ValidateScope(cfg.Scope); err != nil {
		return err
	}
	if err := ValidateTransport(cfg.Transport); err != nil {
		return err
	}
	switch cfg.Transport {
	case "stdio":
		return ValidateStdio(cfg)
	case "sse", "http":
		return ValidateRemote(cfg)
	default:
		// Unreachable: ValidateTransport already rejected unknown values.
		return fmt.Errorf("protocol.Validate: unreachable transport %q: %w", cfg.Transport, ErrInvalidTransport)
	}
}
