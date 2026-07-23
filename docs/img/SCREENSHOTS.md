# Screenshots to capture

Drop PNGs in this folder with the exact names below — the README references them.
Capture from a stand with real data (e.g. the Docker test stand at
`http://localhost:8080`, login `Admin` / `zabbix`, after a `ztc scan`), or any
Zabbix with ztc provisioned. Prefer a light theme, ~1600px wide, trim browser
chrome.

| File | Where in Zabbix | What to show |
|------|-----------------|--------------|
| **`dashboard.png`** _(required — README hero)_ | Monitoring → Dashboards → **Vulners** | The whole dashboard: Problems-by-severity on top, the Hosts/Packages/Bulletins problem lists, and the two graphs (Median CVSS Score + score distribution). This is the headline image. |
| `problems-by-severity.png` | Monitoring → Problems (or the dashboard widget) | The severity breakdown with real counts across **Disaster / High / Average / Warning** — proof it's no longer all "Not classified". |
| `filter-by-host.png` | Monitoring → Problems → Filter → **Tags: `vulners.host` Equals `<host>`** | The filter applied, showing one host's vulnerabilities. Keep the filter row visible so the tag usage is obvious. |
| `graphs.png` _(optional)_ | Dashboard graphs, or Monitoring → Latest data → the two graphs | Close-up of **Median CVSS Score** (trend) + **CVSS Score ratio by hosts** (pie) once some history has accumulated. |
| `problem-detail.png` _(optional)_ | Click a bulletin problem | The problem detail: `[severity] Score … Bulletin = CVE-… on <host>`, the `vulners.host` tag, and the vulners.com link. |

Tips:
- Run a couple of scan cycles first so graphs/problems have data (see the stand README).
- For a fuller dashboard, have both a Linux and a Windows host reporting.
- After adding files, embed the extra ones in the README where useful and commit.
