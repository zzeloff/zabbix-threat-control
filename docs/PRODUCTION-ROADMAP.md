# ztc → Production Roadmap

План приведения Go-версии `ztc` (переписанный `zabbix-threat-control`) к релизу.
Собран из обсуждения (Кир Ермаков / Кирилл Зеленин) + технических хвостов разработки.

**Решение:** код живёт в **старой репе `github.com/…/zabbix-threat-control`** (не в новой),
чтобы сохранить преемственность, звёзды/issue/Stories и привычный для пользователей адрес.

---

## 0. Что уже готово (этот цикл разработки)

- Go-переписывание `ztc`: collect → audit (Vulners) → aggregate → push (zabbix-sender), `provision`, `fix`.
- **Windows**: Smart Audit (софт из реестра → `v4/audit/smart`) + KB (`v3/audit/kb` с per-CVE `details` — CVSS-скоры).
- **Zabbix 6.0 и 7.0**: version-aware клиент (Bearer vs auth-in-body) + version-aware provision.
- **Severity**: 4 бэнда CVSS→priority → «Problems by Severity» распределяется (раньше всё было Not classified).
- **Per-host теги** `vulners.host` (фильтр «хост → его уязвимости», Equals).
- Проверено вживую на стендах 6.0 (`:8081`) и 7.0 (`:8080`), Linux + Windows.

**Продуктовый тезис для README:** зависимость от мажорной версии Zabbix снята; для Windows
используется Smart Audit — работает независимо от версии.

---

## Phase 1 — Консолидация в `zabbix-threat-control`

- [ ] Перенести Go-код (`cmd/ztc`, `internal/`, `deploy/`, `docs/`) в старую репу.
- [ ] **Решение по раскладке:** Go в корне; Python-версию убрать в `legacy/` (тег `v-python-final` на последнем Python-коммите для истории).
- [ ] `go.mod`: module path → путь старой репы; обновить импорты.
- [ ] Сохранить историю (перенос коммитов Go-работы; не терять Python-историю).
- [ ] **go-vulners**: убрать локальный `replace` (см. Phase 2) — иначе Docker-сборка ломается (путь вне контекста).
- [ ] `README`/`AGENTS.md`/`STATUS.md` перевесить на новую структуру.

## Phase 2 — Закрыть технические хвосты

- [ ] **go-vulners `details`**: оформить MR автору (`kidoz/go-vulners`, ветка `feat/kb-details-cvss`, патч готов). До мержа — vendoring или временный форк под своим namespace, чтобы CI/Docker собирались без `replace`-на-локальный-путь.
- [ ] **Backend `/kb details`**: раскатать с beta на **прод `vulners.com`** (сейчас на beta за mTLS). Тогда ztc в проде работает по обычному URL без клиентского сертификата.
- [ ] **Графы дашборда**: портировать из `prepare.py` — median CVSS и «severity/score ratio». (Кирилл: в старой версии график критичности был захардкожен в 0 и никогда не рисовался — в новой должен реально работать.)
- [ ] Ручная перепроверка находок глазами (не только автотесты) — Кирилл.
- [ ] Прочее из TODO: авто-fix вживую (сквозной acknowledge), Windows-ремедиация, Client для 5.0.

## Phase 3 — CI/CD и релизы

- [ ] **Пайплайн** (GitHub Actions или GitLab CI — *решить по площадке репы*): `go build/test/vet`, `gofmt -l`, `golangci-lint`.
- [ ] **goreleaser**: кросс-компиляция бинаря (linux amd64/arm64; при необходимости darwin), checksums, подпись, авто-публикация **GitHub Releases** по тегу.
- [ ] **Docker-образ**: сборка + публикация в GHCR/registry по тегу (multi-arch).
- [ ] Матрица тестов против **Zabbix 6.0 и 7.0** (docker-стенды уже есть в `deploy/docker`).

## Phase 4 — Упаковка и установка (cookbook — главный запрос)

Цель Кира: «одна команда, Enter, 5 секунд — работает». Несколько вариантов под разную инфру.

- [ ] **One-liner `curl … | sh`** (как get-pip): скачивает нужный бинарь под ОС/арх, ставит, поднимает **systemd-сервис** (daemon), интерактивно спрашивает параметры (Vulners API key, Zabbix URL/creds/token), затем `ztc provision --all`. Флаг для неинтерактивного режима (env/CLI).
- [ ] **systemd unit** для `ztc scan --daemon` (+ пример конфига, права, секреты через env).
- [ ] **Ansible role/playbook** — для клиентов с Ansible.
- [ ] **Установка через Zabbix** (wow-вариант): агент уже развёрнут → однократное выполнение команды (script/remote command) поднимает ztc на Zabbix Server **без входа в shell**. ⚠️ Проверить сетевые ограничения и разрешён ли одноразовый run.
- [ ] **Docker / docker-compose quickstart** (стенд уже есть — довести до «prod-ready compose»).
- [ ] Учесть **сетевые ограничения** (закрытый доступ в интернет → офлайн-установка бинаря/образа).

## Phase 5 — Авто-обновление

- [ ] **Проверка версии**: ztc сверяется с GitHub Releases (или version-эндпоинтом), при наличии новой версии — лог + сигнал в Zabbix (item/trigger «доступно обновление»).
- [ ] **Команда самообновления**: `ztc upgrade` (в духе `zabbix-control upgrade`) — скачать новый бинарь, заменить, перезапустить сервис. Если полноценный self-update невозможен в среде клиента — поставлять как отдельную команду/скрипт.

## Phase 6 — README «2026» + документация + скриншоты

- [ ] **Скриншоты** (задача Кирилла): дашборд, Problems by Severity, фильтр `vulners.host`, витрины Bulletins/Packages/Hosts. Потом отдать Claude на сборку README.
- [ ] **README** как продукт-дока 2026: что это / что делает / как работает / зачем нужно; архитектура (диаграмма collect→audit→aggregate→push); quick-start (one-liner); варианты деплоя; конфиг-референс; апгрейд; поддержка версий Zabbix; Windows/Smart Audit.
- [ ] `docs/`: guide, ADR (есть), config reference, migration с Python-версии.
- [ ] Best practices упаковки/структуры репы (по запросу Кира).

---

## Открытые решения (нужен вход команды)

1. **CI-площадка:** GitHub Actions или GitLab CI (зависит от того, где живёт `zabbix-threat-control`).
2. **Таргеты бинаря:** только linux (amd64+arm64) или ещё darwin/windows? (ztc — центральный компонент, обычно рядом с Zabbix Server = Linux.)
3. **Python-версия:** архивировать в `legacy/` + тег, или удалить.
4. **go-vulners:** ждать мерж апстрима vs форк под своим namespace vs vendoring (влияет на сроки CI).
5. **Приоритет install-методов** для v1: скорее всего `curl|sh` + docker первыми, Ansible и via-Zabbix — следом.

## Предлагаемый порядок

**1 (репо) → 2 (хвосты: go-vulners MR + деплой details + графы) → 3 (CI/releases) → 4 (install cookbook) → 5 (auto-update) → 6 (README/скриншоты)**, где 6 идёт параллельно и финализируется после скриншотов.
