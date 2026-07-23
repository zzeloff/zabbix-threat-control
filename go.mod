module github.com/vulnersCom/zabbix-threat-control

go 1.26.5

require (
	github.com/kidoz/go-vulners v1.3.2
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/kidoz/go-vulners => /Users/zeloff/git/go-vulners // TODO(phase2): switch to upstream/forked go-vulners once KB-details PR lands
