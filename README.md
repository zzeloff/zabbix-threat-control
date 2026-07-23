# vulners-zabbix-agent (ztc)

A Go rewrite of [zabbix-threat-control](https://github.com/vulnersCom/zabbix-threat-control):
it turns Zabbix into a vulnerability-management view of your fleet. Hosts report
their inventory through the standard **zabbix-agent2** (no Vulners agent), the
`ztc` binary audits that inventory against [Vulners](https://vulners.com) and
pushes per-host / per-package / per-bulletin findings back into Zabbix.

> **New to Zabbix?** Start with [`docs/guide.md`](docs/guide.md) — the scheme in
> plain language, annotated config examples, and an FAQ.

## What changed vs. the Python version

| Area | Old (Python) | New (Go) |
|------|--------------|----------|
| Vulners API | legacy audit | **v4** `audit/linux`, `audit/smart` (Windows software), **v3** `audit/winaudit` (KB) via [`go-vulners`](https://github.com/kidoz/go-vulners) |
| Host agent | `report.py` deployed per host | **none** — native commands via zabbix-agent2 `UserParameter` (`deploy/agent`) |
| Remediation | `system.run[<cmd>]` / ssh | whitelisted `vulners.fix[<pkg>]` + narrow sudoers (no arbitrary RCE) |
| Platforms | Linux only | Linux **and** Windows |
| Zabbix | 5.0–7.x branches inline | targets **7.x**; a `ZabbixClient` interface leaves room for older versions |
| Delivery | `zabbix_sender` binary | built-in sender protocol |
| Runtime | 4 scripts + deps | single static binary |

## Architecture

```
zabbix-agent2 (UserParameter: vulners.os/version/arch/packages | win.software/win.kb)
        │  (collected into Zabbix items)
        ▼
   ztc scan ──► collect (Zabbix API) ──► audit (Vulners v4/v3) ──► aggregate ──► sender ──► Zabbix
```

Packages (`internal/`): `config`, `model`, `vulners` (Auditor + go-vulners),
`audit` (response→model transforms), `aggregate` (matrices + LLD), `zabbix`
(interface + v7 client + `sender`), `collect`, `provision`, `scan`.

## Build

```sh
make build           # -> bin/ztc
make test            # unit tests
```

## Configure

Copy `config.example.yaml` and edit, or rely on environment variables
(`VULNERS_API_KEY`, `ZABBIX_URL`, `ZABBIX_USER`/`ZABBIX_PASSWORD` or
`ZABBIX_TOKEN`, `ZABBIX_SERVER_FQDN`, `ZABBIX_SERVER_PORT`). Env overrides file.

## Deploy the agent side (no Vulners agent)

Copy the UserParameter snippet to each monitored host and restart the agent:

- Linux: `deploy/agent/linux/vulners.conf` → `/etc/zabbix/zabbix_agent2.d/`
- Windows: `deploy/agent/windows/vulners.conf` → `zabbix_agent2.d\`

These expose fixed keys only — arbitrary `system.run` stays disabled, so the
server cannot execute anything beyond the defined inventory commands.

## Run

```sh
ztc provision --all              # create templates, virtual hosts, dashboard
# link the platform template to each host you want scanned:
#   "Template Vulners OS-Report Linux"   -> Linux hosts
#   "Template Vulners OS-Report Windows" -> Windows hosts
ztc scan --once                  # one audit cycle (for cron / a Zabbix item)
ztc scan --daemon                # run continuously on the configured schedule
ztc scan --daemon --auto-fix     # also remediate problems acknowledged by trusted users
```

## Remediation (fix)

`ztc fix` invokes the whitelisted `vulners.fix[<pkg>]` key on the target's
zabbix-agent2 — no arbitrary `system.run`. Because agent2 kills anything a
UserParameter backgrounds, `vulners.fix` only **queues** the package; a small
root-cron worker on the host performs the upgrade and writes an audit log. See
[`docs/adr/0001-remediation-mechanism.md`](docs/adr/0001-remediation-mechanism.md).

Per host that should be remediable, deploy from `deploy/agent/linux/`:
`vulners.conf` (adds `vulners.fix`), `vulners-fix-worker.sh` → `/usr/local/bin/`,
`vulners-fix.cron` → `/etc/cron.d/`. Also add the ztc host's address to the
agent's `Server=` allowlist (ztc reaches the agent's passive port 10050). No
sudoers is needed — the worker runs as root via cron.

Verify on the host: `cat /var/lib/zabbix/vulners-fix.queue` (pending) and
`tail -f /var/log/vulners-fix.log` (worker `START`/`END rc=…` per package).

```sh
ztc fix --host NAME --all              # upgrade all vulnerable packages on a host
ztc fix --host NAME --package PKG      # upgrade one package on a host
ztc fix --package PKG                  # upgrade PKG on every affected host
ztc fix --host NAME --all --dry-run    # show what would be fixed
```

Automatic remediation (`ztc scan --daemon --auto-fix`) only acts on Zabbix
problems acknowledged by a user listed in `fix.trusted_users`.

## Test stand

See [`deploy/docker/README.md`](deploy/docker/README.md) for a one-command
Zabbix 7.0 + agent + scanner environment.

## Status

See [`STATUS.md`](STATUS.md) for what is implemented and what is deferred.
