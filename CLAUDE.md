# AgentD Protocol Repo

## What This Is
Shared wire protocol types for the AgentD ecosystem. Both the daemon (`agentd`) and relay server (`agentd-relay`) import from this single module to guarantee type identity across the WebSocket JSON contract.

## Build & Test
```bash
go test -race -count=1 ./...
go vet ./...
```

## Architecture
- `envelope.go` — `RelayEnvelope` (encrypted message envelope with trace ID)
- `control.go` — `ControlMessage`, `ControlType` constants (10), all payload structs (9)
- `policy.go` — `PolicyJSON`, `PolicyMatchJSON` (policy rule wire format)
- `trace.go` — `NewTraceID()`, `ValidTraceID()` (W3C trace ID generation)

## Critical Rules

**Zero Dependencies**: Only stdlib imports (`encoding/json`, `time`, `crypto/rand`, `encoding/hex`). NEVER add external dependencies.

**Source of Truth**: This module is the canonical definition of all relay wire types. Both `agentd` and `agentd-relay` use type aliases to re-export these types. **NEVER define wire types directly in the consumer repos' `internal/relay/types.go`.**

**Backward Compatible**: All new fields MUST use `omitempty`. Old clients that don't set new fields must produce identical JSON. New fields must never break existing consumers.

**JSON Tags Are the Contract**: The `json:"..."` tags define the wire format. Changing a tag is a breaking change that affects all 3 repos (daemon, relay, PWA). The TypeScript copy in `agentd-web/src/types/index.ts` must be manually updated to match.

**Roundtrip Tested**: Every type must have a JSON marshal/unmarshal roundtrip test in `protocol_test.go`.

**Trace IDs**: W3C-compatible (32 lowercase hex chars). Uses `crypto/rand` only — never `math/rand`.

**Commits**: `type(scope): subject` format. No AI authorship trailers.

## Adding a New Wire Type

1. Add the struct to the appropriate file (`envelope.go`, `control.go`, or `policy.go`)
2. Add a roundtrip test in `protocol_test.go`
3. Run `go test -race -count=1 ./...`
4. Update the type alias in `agentd/internal/relay/types.go`
5. Update the type alias in `agentd-relay/internal/relay/types.go`
6. Update the TypeScript definition in `agentd-web/src/types/index.ts`

## Full Specification
See `AGENTD_PLAN.md` in the workspace root for the complete wire format specification (Section 5).
