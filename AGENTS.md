# AGENTS.md

Go single binary (`cmd/voki/main.go`) with two roles: `voki server` (controller daemon) and `voki agent` (dials home over mTLS). One binary, both roles.

## Commands

- `make build` — builds `bin/voki` (runs `go mod tidy` and injects version via ldflags; use `VERSION=x make build` to set it)
- `go test ./...` — only `internal/proto` and `internal/targets` have tests
- `make vault` — runs Vault in dev mode (token `123456`) for testing vault integration
- Manual testing without agents: set `host = "localhost"` in a target HCL for local mode; end-to-end needs a controller + agent pair (`voki server` / `voki agent --controller ...`, then `voki peers accept <id>`)

## Layout

- `internal/cli` — cobra commands; each file is one subcommand (`run.go`, `server.go`, `agent.go`, `peers.go`)
- `internal/targets` — HCL parsing, vars, vault, inline funcs (`inline/`: `file()`, `template()`); `types.config.go` is hand-written despite the name
- `internal/controller` / `internal/agent` / `internal/client` / `internal/proto` — daemon side, msgpack over socket (`voki.sock` under `$VOKI_HOME`, default `~/.config/voki`)
- `internal/local` / `internal/remote` — step execution for localhost vs dialed agents

## Gotchas

- Config comes from `voki-config.hcl` or `VOKI_`-prefixed env vars (e.g. `VOKI_VAULT_TOKEN`)
- Default agent port is 7331; target `host` is the agent *name* from `voki peers list`, not an address
- README.md is the full spec for target/task HCL syntax — keep it in sync when changing HCL schema
