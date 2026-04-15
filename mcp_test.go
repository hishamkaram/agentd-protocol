package protocol_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

// TestMCPServerConfigRoundTrip verifies JSON marshal/unmarshal round-tripping
// for MCPServerConfig across all three transports (stdio/sse/http) and both
// scopes (project/user), plus the companion wire types shipped in feature 169.
func TestMCPServerConfigRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.MCPServerConfig
	}{
		{
			name: "stdio project scope enabled with env",
			in: protocol.MCPServerConfig{
				Name:      "filesystem",
				Scope:     protocol.MCPScopeProject,
				Transport: "stdio",
				Enabled:   true,
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
				Env:       map[string]string{"NODE_ENV": "production"},
			},
		},
		{
			name: "stdio user scope disabled no args",
			in: protocol.MCPServerConfig{
				Name:      "github",
				Scope:     protocol.MCPScopeUser,
				Transport: "stdio",
				Enabled:   false,
				Command:   "/usr/local/bin/mcp-github",
			},
		},
		{
			name: "sse https project scope with headers",
			in: protocol.MCPServerConfig{
				Name:      "docs",
				Scope:     protocol.MCPScopeProject,
				Transport: "sse",
				Enabled:   true,
				URL:       "https://code.claude.com/docs/mcp",
				Headers:   map[string]string{"Authorization": "Bearer xxx"},
			},
		},
		{
			name: "http user scope localhost",
			in: protocol.MCPServerConfig{
				Name:      "local-http",
				Scope:     protocol.MCPScopeUser,
				Transport: "http",
				Enabled:   true,
				URL:       "http://127.0.0.1:8080/mcp",
			},
		},
		{
			name: "http project scope https enabled",
			in: protocol.MCPServerConfig{
				Name:      "remote-api",
				Scope:     protocol.MCPScopeProject,
				Transport: "http",
				Enabled:   true,
				URL:       "https://api.example.com/mcp",
				Headers:   map[string]string{"X-API-Key": "abc"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var out protocol.MCPServerConfig
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(tt.in, out) {
				t.Errorf("roundtrip mismatch:\n in:  %+v\n out: %+v", tt.in, out)
			}
		})
	}
}

func TestMCPServerConfigOmitempty(t *testing.T) {
	t.Parallel()

	// stdio server: URL and Headers must be omitted.
	stdio := protocol.MCPServerConfig{
		Name:      "fs",
		Scope:     protocol.MCPScopeProject,
		Transport: "stdio",
		Enabled:   true,
		Command:   "npx",
	}
	data, err := json.Marshal(stdio)
	if err != nil {
		t.Fatalf("marshal stdio: %v", err)
	}
	s := string(data)
	for _, banned := range []string{`"url"`, `"headers"`} {
		if strings.Contains(s, banned) {
			t.Errorf("stdio config should omit %s, got: %s", banned, s)
		}
	}

	// sse server: Command, Args, Env must be omitted.
	sse := protocol.MCPServerConfig{
		Name:      "docs",
		Scope:     protocol.MCPScopeUser,
		Transport: "sse",
		Enabled:   true,
		URL:       "https://example.com",
	}
	data, err = json.Marshal(sse)
	if err != nil {
		t.Fatalf("marshal sse: %v", err)
	}
	s = string(data)
	for _, banned := range []string{`"command"`, `"args"`, `"env"`} {
		if strings.Contains(s, banned) {
			t.Errorf("sse config should omit %s, got: %s", banned, s)
		}
	}
}

func TestMCPServerStatusEntryRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.MCPServerStatusEntry
	}{
		{
			name: "connected with tools",
			in: protocol.MCPServerStatusEntry{
				Name:      "filesystem",
				Status:    "connected",
				ToolCount: 12,
			},
		},
		{
			name: "failed with message",
			in: protocol.MCPServerStatusEntry{
				Name:    "broken",
				Status:  "failed",
				Message: "connection refused",
			},
		},
		{
			name: "shadowed by user",
			in: protocol.MCPServerStatusEntry{
				Name:       "github",
				Status:     "disabled",
				ShadowedBy: protocol.MCPScopeUser,
			},
		},
		{
			name: "disconnected",
			in: protocol.MCPServerStatusEntry{
				Name:   "api",
				Status: "disconnected",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var out protocol.MCPServerStatusEntry
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(tt.in, out) {
				t.Errorf("mismatch:\n in:  %+v\n out: %+v", tt.in, out)
			}
		})
	}
}

func TestMCPListPayloadRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.MCPListPayload
	}{
		{name: "both scopes no session", in: protocol.MCPListPayload{}},
		{name: "project scope only", in: protocol.MCPListPayload{Scope: protocol.MCPScopeProject, SessionID: "sess-1"}},
		{name: "user scope only", in: protocol.MCPListPayload{Scope: protocol.MCPScopeUser}},
		// Feature 170: RequestID field round-trips under Unmarshal.
		{name: "with request_id", in: protocol.MCPListPayload{RequestID: "req-abc", Scope: protocol.MCPScopeUser}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			var out protocol.MCPListPayload
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if tt.in != out {
				t.Errorf("mismatch: %+v vs %+v", tt.in, out)
			}
		})
	}
}

func TestMCPListResponseRoundTrip(t *testing.T) {
	t.Parallel()

	// Feature 170: RequestID field is echoed back from the originating request.
	in := protocol.MCPListResponse{
		RequestID: "req-list-42",
		Servers: []protocol.MCPServerConfig{
			{Name: "a", Scope: protocol.MCPScopeProject, Transport: "stdio", Enabled: true, Command: "a"},
			{Name: "b", Scope: protocol.MCPScopeUser, Transport: "http", Enabled: true, URL: "https://b.example.com"},
		},
		Status: []protocol.MCPServerStatusEntry{
			{Name: "a", Status: "connected", ToolCount: 4},
			{Name: "b", Status: "failed", Message: "timeout"},
		},
		ProjectWorkdir: "/tmp/proj",
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out protocol.MCPListResponse
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("mismatch:\n in:  %+v\n out: %+v", in, out)
	}
	if out.RequestID != "req-list-42" {
		t.Errorf("RequestID not round-tripped: got %q, want %q", out.RequestID, "req-list-42")
	}
}

// TestMCPListResponseRequestIDOmitempty verifies the request_id field is
// omitted from the wire when unset, so legacy PWAs that do not send
// request_id see identical JSON output (feature 170 backward-compat).
func TestMCPListResponseRequestIDOmitempty(t *testing.T) {
	t.Parallel()

	in := protocol.MCPListResponse{
		Servers: []protocol.MCPServerConfig{},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); strings.Contains(got, "request_id") {
		t.Errorf("unexpected request_id in empty response JSON: %s", got)
	}
}

func TestMCPMutationPayloadRoundTrip(t *testing.T) {
	t.Parallel()

	// Feature 170: RequestID on the request so the daemon can echo it back.
	in := protocol.MCPMutationPayload{
		RequestID: "req-mut-42",
		SessionID: "sess-42",
		Server: protocol.MCPServerConfig{
			Name:      "fs",
			Scope:     protocol.MCPScopeProject,
			Transport: "stdio",
			Enabled:   true,
			Command:   "npx",
			Args:      []string{"-y", "@modelcontextprotocol/server-filesystem"},
		},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out protocol.MCPMutationPayload
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("mismatch:\n in:  %+v\n out: %+v", in, out)
	}
}

func TestMCPMutationResponseRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.MCPMutationResponse
	}{
		{
			name: "ok applied to live with request_id (feature 170)",
			in: protocol.MCPMutationResponse{
				RequestID:     "req-mut-ok-1",
				OK:            true,
				AppliedToLive: true,
				LiveStatus: []protocol.MCPServerStatusEntry{
					{Name: "fs", Status: "connected", ToolCount: 5},
				},
				Server: &protocol.MCPServerConfig{
					Name: "fs", Scope: protocol.MCPScopeProject, Transport: "stdio",
					Enabled: true, Command: "npx",
				},
			},
		},
		{
			name: "error validation with request_id (feature 170)",
			in: protocol.MCPMutationResponse{
				RequestID: "req-mut-err-1",
				OK:        false,
				Error:     "invalid name",
				ErrorCode: "invalid_name",
			},
		},
		{
			name: "ok persisted no live session no request_id",
			in: protocol.MCPMutationResponse{
				OK:            true,
				AppliedToLive: false,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			var out protocol.MCPMutationResponse
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.in, out) {
				t.Errorf("mismatch:\n in:  %+v\n out: %+v", tt.in, out)
			}
		})
	}
}

func TestMCPRemoveTogglReconnectRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("remove", func(t *testing.T) {
		t.Parallel()
		// Feature 170: RequestID on every request payload.
		in := protocol.MCPRemovePayload{
			RequestID: "req-rm-1",
			SessionID: "sess-1",
			Scope:     protocol.MCPScopeProject,
			Name:      "fs",
		}
		data, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		var out protocol.MCPRemovePayload
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}
		if in != out {
			t.Errorf("mismatch: %+v vs %+v", in, out)
		}
	})

	t.Run("toggle", func(t *testing.T) {
		t.Parallel()
		in := protocol.MCPTogglePayload{
			RequestID: "req-tg-1",
			SessionID: "sess-1",
			Scope:     protocol.MCPScopeUser,
			Name:      "api",
			Enabled:   true,
		}
		data, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		var out protocol.MCPTogglePayload
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}
		if in != out {
			t.Errorf("mismatch: %+v vs %+v", in, out)
		}
	})

	t.Run("reconnect", func(t *testing.T) {
		t.Parallel()
		in := protocol.MCPReconnectPayload{RequestID: "req-rc-1", SessionID: "sess-1", Name: "fs"}
		data, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		var out protocol.MCPReconnectPayload
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}
		if in != out {
			t.Errorf("mismatch: %+v vs %+v", in, out)
		}
	})
}

func TestMCPServersChangedPayloadRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []protocol.MCPServersChangedPayload{
		{Scope: protocol.MCPScopeProject, SessionID: "sess-1", Action: "add", ServerName: "fs"},
		{Scope: protocol.MCPScopeUser, Action: "remove", ServerName: "gh"},
		{Scope: protocol.MCPScopeProject, SessionID: "sess-2", Action: "toggle", ServerName: "api"},
		{Scope: protocol.MCPScopeUser, Action: "update", ServerName: "docs"},
	}
	for _, in := range tests {
		in := in
		t.Run(in.Action+"_"+string(in.Scope), func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(in)
			if err != nil {
				t.Fatal(err)
			}
			var out protocol.MCPServersChangedPayload
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if in != out {
				t.Errorf("mismatch: %+v vs %+v", in, out)
			}
		})
	}
}

// TestMCPServerConfigValidation covers every validation rule in
// data-model.md §Validation rules.
func TestMCPServerConfigValidation(t *testing.T) {
	t.Parallel()

	base := func() protocol.MCPServerConfig {
		return protocol.MCPServerConfig{
			Name:      "ok-name",
			Scope:     protocol.MCPScopeProject,
			Transport: "stdio",
			Enabled:   true,
			Command:   "npx",
			Args:      []string{"-y", "pkg"},
		}
	}

	tests := []struct {
		name    string
		mutate  func(*protocol.MCPServerConfig)
		wantErr error
	}{
		// Valid baseline.
		{name: "valid stdio", mutate: func(c *protocol.MCPServerConfig) {}, wantErr: nil},
		{
			name: "valid sse https",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeUser, Transport: "sse", Enabled: true,
					URL: "https://example.com/sse",
				}
			},
			wantErr: nil,
		},
		{
			name: "valid http localhost",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "l", Scope: protocol.MCPScopeProject, Transport: "http", Enabled: true,
					URL: "http://localhost:8080/mcp",
				}
			},
			wantErr: nil,
		},
		{
			name: "valid http 127.0.0.1",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "l", Scope: protocol.MCPScopeProject, Transport: "http", Enabled: true,
					URL: "http://127.0.0.1/mcp",
				}
			},
			wantErr: nil,
		},
		{
			name: "valid http ipv6 loopback",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "l", Scope: protocol.MCPScopeProject, Transport: "http", Enabled: true,
					URL: "http://[::1]:8080/mcp",
				}
			},
			wantErr: nil,
		},

		// Name rules.
		{
			name:    "name empty",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = "" },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name leading hyphen",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = "-bad" },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name leading dot",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = ".hidden" },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name contains slash",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = "foo/bar" },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name contains backslash",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = `foo\bar` },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name dotdot",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = ".." },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name reserved .mcp.json",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = ".mcp.json" },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name too long",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = strings.Repeat("a", 65) },
			wantErr: protocol.ErrInvalidName,
		},
		{
			name:    "name invalid unicode",
			mutate:  func(c *protocol.MCPServerConfig) { c.Name = "naïve" },
			wantErr: protocol.ErrInvalidName,
		},

		// Transport rules.
		{
			name:    "transport empty",
			mutate:  func(c *protocol.MCPServerConfig) { c.Transport = "" },
			wantErr: protocol.ErrInvalidTransport,
		},
		{
			name:    "transport sdk rejected",
			mutate:  func(c *protocol.MCPServerConfig) { c.Transport = "sdk" },
			wantErr: protocol.ErrInvalidTransport,
		},
		{
			name:    "transport unknown",
			mutate:  func(c *protocol.MCPServerConfig) { c.Transport = "websocket" },
			wantErr: protocol.ErrInvalidTransport,
		},

		// Scope rules.
		{
			name:    "scope empty",
			mutate:  func(c *protocol.MCPServerConfig) { c.Scope = "" },
			wantErr: protocol.ErrInvalidScope,
		},
		{
			name:    "scope local rejected",
			mutate:  func(c *protocol.MCPServerConfig) { c.Scope = "local" },
			wantErr: protocol.ErrInvalidScope,
		},

		// Stdio command rules.
		{
			name:    "stdio command empty",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "" },
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name: "stdio command too long",
			mutate: func(c *protocol.MCPServerConfig) {
				c.Command = strings.Repeat("x", 2049)
			},
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name:    "stdio command shell pipe",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "sh -c 'echo | cat'" },
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name:    "stdio command backtick",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "echo `whoami`" },
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name:    "stdio command dollar",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "echo $HOME" },
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name:    "stdio command semicolon",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "a;b" },
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name:    "stdio command ampersand",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "a&b" },
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name:    "stdio command redirect in",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "a<b" },
			wantErr: protocol.ErrInvalidCommand,
		},
		{
			name:    "stdio command redirect out",
			mutate:  func(c *protocol.MCPServerConfig) { c.Command = "a>b" },
			wantErr: protocol.ErrInvalidCommand,
		},

		// Stdio args rules.
		{
			name: "stdio args too many",
			mutate: func(c *protocol.MCPServerConfig) {
				args := make([]string, 65)
				for i := range args {
					args[i] = "a"
				}
				c.Args = args
			},
			wantErr: protocol.ErrInvalidArgs,
		},
		{
			name: "stdio arg too long",
			mutate: func(c *protocol.MCPServerConfig) {
				c.Args = []string{strings.Repeat("a", 1025)}
			},
			wantErr: protocol.ErrInvalidArgs,
		},

		// Stdio env rules.
		{
			name: "stdio env too many keys",
			mutate: func(c *protocol.MCPServerConfig) {
				env := make(map[string]string, 33)
				for i := 0; i < 33; i++ {
					env["K_"+strings.Repeat("X", i+1)] = "v"
				}
				c.Env = env
			},
			wantErr: protocol.ErrInvalidEnv,
		},
		{
			name: "stdio env key lowercase",
			mutate: func(c *protocol.MCPServerConfig) {
				c.Env = map[string]string{"path": "/bin"}
			},
			wantErr: protocol.ErrInvalidEnv,
		},
		{
			name: "stdio env key leading digit",
			mutate: func(c *protocol.MCPServerConfig) {
				c.Env = map[string]string{"1BAD": "v"}
			},
			wantErr: protocol.ErrInvalidEnv,
		},
		{
			name: "stdio env value too long",
			mutate: func(c *protocol.MCPServerConfig) {
				c.Env = map[string]string{"OK": strings.Repeat("a", 4097)}
			},
			wantErr: protocol.ErrInvalidEnv,
		},
		{
			name: "stdio env valid",
			mutate: func(c *protocol.MCPServerConfig) {
				c.Env = map[string]string{"NODE_ENV": "prod", "LOG_LEVEL": "info", "_UNDERSCORE": "ok"}
			},
			wantErr: nil,
		},

		// Transport exclusivity: stdio with URL/Headers is invalid.
		{
			name: "stdio with url",
			mutate: func(c *protocol.MCPServerConfig) {
				c.URL = "https://example.com"
			},
			wantErr: protocol.ErrInvalidTransportFields,
		},
		{
			name: "stdio with headers",
			mutate: func(c *protocol.MCPServerConfig) {
				c.Headers = map[string]string{"X-Key": "v"}
			},
			wantErr: protocol.ErrInvalidTransportFields,
		},

		// Remote URL rules.
		{
			name: "sse missing url",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeProject, Transport: "sse", Enabled: true,
				}
			},
			wantErr: protocol.ErrInvalidURL,
		},
		{
			name: "sse http non-loopback",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeProject, Transport: "sse", Enabled: true,
					URL: "http://example.com/sse",
				}
			},
			wantErr: protocol.ErrInvalidURL,
		},
		{
			name: "sse unparseable",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeProject, Transport: "sse", Enabled: true,
					URL: "://bad",
				}
			},
			wantErr: protocol.ErrInvalidURL,
		},
		{
			name: "sse url too long",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeProject, Transport: "sse", Enabled: true,
					URL: "https://example.com/" + strings.Repeat("a", 2100),
				}
			},
			wantErr: protocol.ErrInvalidURL,
		},
		{
			name: "http file scheme rejected",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeProject, Transport: "http", Enabled: true,
					URL: "file:///etc/passwd",
				}
			},
			wantErr: protocol.ErrInvalidURL,
		},

		// Remote header rules.
		{
			name: "sse headers too many",
			mutate: func(c *protocol.MCPServerConfig) {
				h := make(map[string]string, 33)
				for i := 0; i < 33; i++ {
					h["X-"+strings.Repeat("a", i+1)] = "v"
				}
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeUser, Transport: "sse", Enabled: true,
					URL: "https://example.com", Headers: h,
				}
			},
			wantErr: protocol.ErrInvalidHeaders,
		},
		{
			name: "sse header key invalid char",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeUser, Transport: "sse", Enabled: true,
					URL:     "https://example.com",
					Headers: map[string]string{"X-Bad Header": "v"},
				}
			},
			wantErr: protocol.ErrInvalidHeaders,
		},
		{
			name: "sse header value too long",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeUser, Transport: "sse", Enabled: true,
					URL:     "https://example.com",
					Headers: map[string]string{"X-OK": strings.Repeat("v", 4097)},
				}
			},
			wantErr: protocol.ErrInvalidHeaders,
		},
		{
			name: "sse headers valid rfc7230",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeUser, Transport: "sse", Enabled: true,
					URL:     "https://example.com",
					Headers: map[string]string{"X-API-Key": "abc", "Authorization": "Bearer xxx"},
				}
			},
			wantErr: nil,
		},

		// Transport exclusivity: sse with stdio fields.
		{
			name: "sse with command",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeUser, Transport: "sse", Enabled: true,
					URL: "https://example.com", Command: "/bin/true",
				}
			},
			wantErr: protocol.ErrInvalidTransportFields,
		},
		{
			name: "http with args",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeProject, Transport: "http", Enabled: true,
					URL: "https://example.com", Args: []string{"--oops"},
				}
			},
			wantErr: protocol.ErrInvalidTransportFields,
		},
		{
			name: "http with env",
			mutate: func(c *protocol.MCPServerConfig) {
				*c = protocol.MCPServerConfig{
					Name: "s", Scope: protocol.MCPScopeProject, Transport: "http", Enabled: true,
					URL: "https://example.com", Env: map[string]string{"X": "y"},
				}
			},
			wantErr: protocol.ErrInvalidTransportFields,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := base()
			tt.mutate(&cfg)
			err := protocol.Validate(&cfg)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate(%+v): unexpected error: %v", cfg, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate(%+v): expected error %v, got nil", cfg, tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate(%+v): error = %v, want %v", cfg, err, tt.wantErr)
			}
		})
	}
}
