<h1 align="center">Pluggent</h1>
<p align="center">Building an all-plugin Agent Harness from scratch in Go.</p>
<p align="center"><a href="README.md">简体中文</a> · <a href="README.en.md">English</a></p>

> Current stage: the plugin kernel and configuration-driven composition run end to end. The agent loop, models, tools, and skills are still planned.

## Quick start

Use the Go version declared in `go.mod`. From the repository root:

```bash
make run
```

Output:

```text
pluggent dev
```

The Makefile injects the build version. Use another value to exercise the Provider → Service → Consumer path:

```bash
make run VERSION=0.1.0
# pluggent 0.1.0
```

A local `.env` file is optional. To point Pluggent at another YAML file, copy the example and edit its path:

```bash
cp .env.example .env
```

## How it works today

1. The process reads `PLUGGENT_CONFIG` to locate the YAML file, or uses `config/pluggent.yaml` by default.
2. The composition root creates plugins from registered factories in YAML order.
3. `Host` gives each plugin its own `Scope`; those scopes share one service registry.
4. `RuntimeInfoProviderPlugin` provides `RuntimeInfoService`. `VersionReporterPlugin` resolves it and prints the version.
5. If startup fails, Host rolls back loaded plugins. Normal shutdown cleans them up in reverse startup order.

```text
config/pluggent.yaml
        │ selects plugins in order
        ▼
      Host ─────────────── shared service registry
       ├─ RuntimeInfoProviderPlugin ─ Scope A ─ Provide(Info) ─┐
       └─ VersionReporterPlugin    ─ Scope B ─ Resolve(Info) ◀─┘
```

| Concept | Responsibility | Code |
|---|---|---|
| `Plugin` | Plugin ID and `Apply` method | [`internal/plugin/plugin.go`](internal/plugin/plugin.go) |
| `Host` | Startup, shared services, rollback, and shutdown | [`internal/plugin/host.go`](internal/plugin/host.go) |
| `Scope` | Cleanup owned by one plugin | [`internal/plugin/scope.go`](internal/plugin/scope.go) |
| `Service[T]` | Typed service provision and resolution | [`internal/plugin/service.go`](internal/plugin/service.go) |
| Composition root | Maps YAML plugin IDs to concrete plugins | [`internal/app/composition.go`](internal/app/composition.go) |

## Configure plugins

The default configuration lives in [`config/pluggent.yaml`](config/pluggent.yaml):

```yaml
plugins:
  - id: runtime-info-provider
  - id: version-reporter
```

Order matters: the provider starts before its consumer resolves the service. Set `disabled: true` to disable an entry. Unknown enabled IDs, duplicate IDs, unknown YAML fields, and missing required services fail startup.

`.env.example` documents the configuration path:

```dotenv
PLUGGENT_CONFIG=config/pluggent.yaml
```

Path precedence is **process environment → `.env` → code default**. You can select a file for one run:

```bash
PLUGGENT_CONFIG=/path/to/pluggent.yaml make run
```

YAML currently selects only plugins compiled into the factory map in [`internal/app/composition.go`](internal/app/composition.go). Loading arbitrary Go code and hot reload are not implemented.

## Development commands

| Command | Purpose |
|---|---|
| `make run` | Run from source |
| `make check` | Run `go vet` and unit tests |
| `make test-race` | Run tests with the race detector |
| `make build` | Build `bin/pluggent` |
| `make fmt` | Format Go source |
| `make help` | List all commands |

## Next steps

The plugin kernel is the foundation of the Harness. Planned work includes a model service and provider registry, an agent loop, tools and skills, durable sessions, context assembly and compaction, and interchangeable CLI, TUI, and Web entry points.

This is also a project for learning Go and the engineering behind Agent Harnesses. The sections above describe running code; this section describes future work.
