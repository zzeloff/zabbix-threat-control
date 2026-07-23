# AGENTS.md

Guidance for AI coding agents (and new contributors) working in this repository.

## What this is

`ztc` — a Go service that turns Zabbix into a vulnerability-management view. Hosts
report their package inventory via stock **zabbix-agent2 UserParameters**; the
central `ztc` binary reads that inventory from the Zabbix API, audits it against
the **Vulners API**, and pushes findings back into Zabbix. It optionally remediates
by upgrading vulnerable packages. It is a Go rewrite of the Python
`zabbix-threat-control`.

Read [`docs/guide.md`](docs/guide.md) for the architecture in plain language.

## Layout

```
cmd/ztc/            CLI entrypoint (subcommands: scan, provision, fix, version)
internal/
  config/           YAML + env config (config.Config)
  model/            source-neutral domain types (Host, HostResult, Package, Bulletin)
  vulners/          Auditor interface + go-vulners wrapper + Mock
  audit/            transform Vulners responses -> model; per-OS fix commands
  aggregate/        build Zabbix trapper items + LLD payloads (scan.py port)
  zabbix/           ZabbixClient interface, v7 JSON-RPC impl, factory, Mock
    sender/         zabbix-sender (trapper) protocol client
    agentget/       passive zabbix-agent protocol client (zabbix_get equivalent)
  collect/          read host inventory from Zabbix items -> model.Host
  provision/        create Zabbix entities (templates, hosts, triggers, dashboard)
  fix/              remediation: manual CLI + acknowledge-driven auto-fix
  scan/             orchestrates collect -> audit -> aggregate -> push
deploy/
  agent/linux/      zabbix-agent2 UserParameter snippet + fix worker + cron
  docker/           test stand (docker compose)
docs/               guide + ADRs
```

## Build / test / run

```sh
go build ./...          # compile
go test ./...           # unit tests (no network required)
go vet ./...            # static checks
make build              # -> bin/ztc
gofmt -l .              # must print nothing
```

Test stand (Docker): see [`deploy/docker/README.md`](deploy/docker/README.md).

Dependencies: only `github.com/kidoz/go-vulners` and `gopkg.in/yaml.v3`. Prefer the
standard library; do not add dependencies without a strong reason.

## Conventions

- **Interfaces are the seams.** External systems sit behind interfaces with test
  doubles: `vulners.Auditor` (`vulners.Mock`) and `zabbix.Client` (`zabbix.Mock`).
  Add new behaviour behind these, not with direct calls.
- **Zabbix version support.** Only `zabbix/v7.go` exists; `zabbix.New` (factory)
  selects it. Add older versions as new `Client` implementations — keep
  version-specific quirks inside the implementation, never in business logic.
- **Pure functions + table-driven tests.** `audit` and `aggregate` are pure
  transforms; cover changes with table-driven tests. Follow TDD where practical.
- **Config.** All tunables go through `internal/config`; secrets via env
  (`VULNERS_API_KEY`, `ZABBIX_*`). `config.Duration` parses `"30s"`/`"24h"`.
- **Style.** Standard Go; run `gofmt`. Match surrounding code; keep files focused.
- Comments explain *why*, not *what*. Keep them sparse and accurate.

## How to extend (common tasks)

- **New audit source / Vulners endpoint:** add a method to `vulners.Auditor` (+
  `Mock`), wrap the SDK in `vulners.Client`, transform in `internal/audit`.
- **Support another Zabbix version:** implement `zabbix.Client` in a new file,
  wire it into `zabbix.New`.
- **New collected field:** add a UserParameter key in `deploy/agent/linux/`,
  read it in `internal/collect`, thread it through `model.Host`.

## Design decisions

Remediation runs vulnerable-package upgrades through a whitelisted
`vulners.fix[<pkg>]` UserParameter + a host-side spool worker (no arbitrary
`system.run`). Rationale and rejected alternatives:
[`docs/adr/0001-remediation-mechanism.md`](docs/adr/0001-remediation-mechanism.md).

## Before you finish

- `go test ./...` and `go vet ./...` pass; `gofmt -l .` is clean.
- Update docs when behaviour or config changes (README, `docs/guide.md`, ADRs).
- Do not commit secrets or a real `config.yaml`; `deploy/docker/.env` is gitignored.
