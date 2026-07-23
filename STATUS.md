# Статус реализации

Дата: 2026-07-18. Итог первого этапа переписывания `zabbix-threat-control` на Go.

## ✅ Сделано

### Ядро (сбор → аудит → агрегация → пуш)
- **`internal/config`** — загрузка YAML + override из env, дефолты как в `config.py`,
  кастомный тип `Duration` (`30s`/`24h`), валидация секретов. Тесты.
- **`internal/model`** — единая доменная модель (Host / HostResult / Package / Bulletin),
  нейтральная к источнику (Linux/Windows/KB).
- **`internal/vulners`** — интерфейс `Auditor` + обёртка над `go-vulners v1.3.2`:
  `LinuxAudit` (v4/audit/linux, с `cvelist_metrics` для CVSS), `WindowsSoftwareAudit`
  (v4/audit/smart), `WindowsKBAudit` (v3 KB). Мок для тестов.
- **`internal/audit`** — трансформеры ответов vulners → `model` (порт
  `_transform_linux_audit_response` и `_get_fix_command` на v4-формат), маршрутизация
  по платформе. Table-driven тесты (Linux/Windows/fix-команды/роутинг).
- **`internal/aggregate`** — порт `write_*_data`: матрицы hosts/packages/bulletins/
  statistics, агрегаты (median/mean/max/min, гистограмма 0–10), LLD-JSON с макросами
  `{#H.*}`/`{#PKG.*}`/`{#BULLETIN.*}`. Тесты (включая пустой ввод).
- **`internal/scan`** — оркестратор цикла (collect→audit→aggregate→push, LLD→delay→data),
  режимы `--once` и `--daemon`. Тесты на моках (две партии пуша, пустой список хостов).

### Zabbix (seam под версии)
- **`internal/zabbix`** — интерфейс `ZabbixClient` (read-путь типизирован, write-путь через
  `Call`), фабрика по версии, реализация **v7** (JSON-RPC, аутентификация `Authorization:
  Bearer` для API-токена и сессии `user.login`), мок.
- **`internal/zabbix/sender`** — клиент протокола zabbix-sender (`ZBXD\x01`+len+JSON).
  Тест round-trip против фейкового сервера.
- **`internal/collect`** — чтение inventory из Zabbix items, детект платформы,
  нормализация (`ol`→`oraclelinux`), фильтр валидности (порт `verify_os_data`). Тесты.

### Provisioning (аналог prepare.py, Zabbix 7.x)
- **`internal/provision`** — host group, template group, **два шаблона сбора**
  (Linux и Windows — чтобы хост получал только свои ключи, без «Unknown metric»),
  4 витрин-хоста (LLD + item-прототипы + trigger-прототипы), статистические
  trapper-item'ы, дашборд. Тесты на моках.

### Отказ от агента vulners
- **`deploy/agent/linux/vulners.conf`** — UserParameter на нативных командах
  (`/etc/os-release`, `dpkg-query`/`rpm -qa`/`apk info`, `uname`). Проверено на живом
  zabbix-agent2 (alpine): `vulners.os`, `vulners.version`, `vulners.packages` отдаются.
- **`deploy/agent/windows/vulners.conf`** — UserParameter (PowerShell): софт из реестра
  (Uninstall) → smart-audit, KB → winaudit, OS caption/version.
- Произвольный `system.run` не включается — сервер вызывает только фиксированные ключи.

### Ремедиация (fix) — реализована (spool + host-worker)
- **`internal/fix`** — инициация: whitelisted `vulners.fix[<pkg>]` через пассивный
  agent-протокол (`internal/zabbix/agentget`, аналог zabbix_get); **без произвольного
  `system.run`** (см. `docs/adr/0001-remediation-mechanism.md`).
- **Исполнение — spool + host-worker:** `vulners.fix[pkg]` кладёт пакет в очередь-файл
  (agent2 убивает фоновые задачи UserParameter, поэтому апгрейд там выполнять нельзя);
  root-cron воркер `vulners-fix-worker.sh` забирает очередь, обновляет пакеты и пишет
  аудит-лог `/var/log/vulners-fix.log`. sudoers не нужен (воркер под root).
- Режимы: ручной CLI (`ztc fix --host X --all|--package Y`, `--package Y` на всех
  затронутых хостах, `--dry-run`) и авто (`ztc scan --daemon --auto-fix`) по acknowledge
  доверенным пользователем (`fix.trusted_users`). Гранулярность: пакет и весь хост.
- Теги триггеров (`vulners.target`/`vulners.package`) в provision — для авто-сопоставления
  событий с хостами/пакетами без парсинга текста. Тесты на моках + apk-имена.
- **`deploy/agent/linux/`**: `vulners.conf` (+`vulners.fix`→очередь), `vulners-fix-worker.sh`,
  `vulners-fix.cron`.

### CLI и упаковка
- **`cmd/ztc`** — подкоманды `scan` (`--once`/`--daemon`/`--auto-fix`/`--nopush`/`--limit`/
  `--push-delay`), `provision` (`--all`/`--template`/`--vhosts`/`--dashboard`), `fix`, `version`.
- **Dockerfile** (multi-stage, статический бинарь), **Makefile**, `config.example.yaml`.

### Тестовый стенд (Docker)
- **`deploy/docker`** — compose: Zabbix 7.0 (server+web+Postgres) + zabbix-agent2 + ztc
  (сервис ztc — постоянный демон `scan --daemon`, `restart: unless-stopped`, `ZTC_SCHEDULE`).

## 🧪 Проверено вживую (на стенде Zabbix 7.0.28 + локальный vulners)
- `ztc provision --all` создал шаблон, 3 витрин-хоста, статистику и дашборд —
  подтверждён v7 JSON-RPC клиент, аутентификация и весь write-путь.
- **Полный цикл `ztc scan`**: collect (Zabbix API) → audit (vulners v4) → aggregate →
  sender → Zabbix. Данные легли в витрины: `vulners.TotalHosts=1`, LLD создал
  прототип-item `vulners.hosts[<id>]`. Sender-путь подтверждён.
- **Ремедиация `ztc fix`** (spool+worker): `ztc fix --package musl` → агент поставил в
  очередь (`response=queued`) → root-воркер выполнил `apk upgrade musl` → аудит-лог
  `/var/log/vulners-fix.log`: `START musl` / `OK: 22.5 MiB in 40 packages` / `END rc=0`.
  Подтверждён agentget против живого агента, обрезка apk-имён, требование Server-allowlist.
- zabbix-agent2 отдаёт `vulners.*` ключи (в т.ч. `vulners.fix`) из смонтированного сниппета.
- `go test ./...` — все пакеты зелёные; `go vet` чист.

## 🐞 Найден и исправлен баг бэкенда vulners
- `v4/audit/linux` падал с **HTTP 500** на любом alpine-аудите: `LinuxAuditApplicableAdvisory.
  operator` был `Literal["lt","le"]`, а `_linux.py` эмитит `"eq"` для точных fixed-версий
  (alpine secfixes) → `ValidationError`. Исправлено в `~/git/vulners/core`
  (`services/audit_service/_types.py`: добавлен `"eq"`); alpine-аудит теперь отдаёт данные.

## ⛔ Осталось / вне текущего этапа

1. **Авто-fix (`--auto-fix`) вживую.** Ядро и распознавание acknowledge покрыты unit-тестами;
   сквозной прогон «подтверждение доверенным пользователем в UI → авто-ремедиация» на живом
   Zabbix не гонялся (нужен реальный acknowledge). Ручной `ztc fix` проверен end-to-end.
2. **Windows-ремедиация.** `vulners.fix` реализован для Linux (apt/yum/apk/zypper); для
   Windows обновление софта/KB — отдельный сложный механизм, не покрыт.
3. **Zabbix action → `ztc fix` (вариант B).** Отложен; при необходимости — тонкая обёртка
   над тем же `ztc fix`.
4. **Реализации `ZabbixClient` для 6.0/5.0.** Есть seam (интерфейс + фабрика по версии),
   но реализована только 7.x. Фабрика предупреждает, если детектит <7.
5. **Windows-нюансы аудита:**
   - `KBAudit` требует точное имя ОС из бюллетеней vulners (напр. «Windows 10 Version 22H2
     for x64-based Systems»); сейчас передаётся `Win32_OperatingSystem.Caption` — вероятно
     нужен маппинг.
   - smart-audit не возвращает CVSS напрямую — score берётся из `ai_score` (может быть 0);
     для точных баллов Windows-софта нужна доп. обогащающая ручка.
   - Windows Go-коллектор (fallback C) не понадобился — хватило UserParameter.
6. **RHEL-семейство: `VERSION_ID`.** Передаётся как есть (напр. «8.9»); при необходимости
   обрезать до мажора для части дистрибутивов.
7. **Provisioning: графы и `--utils`.** Дашборд создаётся с problem-виджетами; графы
   median/score-ratio из `prepare.py` и предстартовая проверка `--utils` (доступность
   агента/sender) пока не портированы. Повторный запуск provision не пересоздаёт уже
   существующие объекты (оставляет как есть) — авто-backup/rename из `prepare.py` упрощён.
8. **Импорт готового Zabbix-шаблона (`templates/*.yaml`)** как альтернатива provision —
   не сделан.
9. **Наблюдаемость/CI.** Логи через `slog`; метрики, `golangci-lint`, CI-пайплайн и
   релизные артефакты (goreleaser) не настроены.
