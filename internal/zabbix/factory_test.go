package zabbix

import "testing"

func TestUseBodyAuth(t *testing.T) {
	cases := map[string]bool{
		"7.0.4":   false, // 7.x -> Bearer
		"7.2.0":   false,
		"6.4.0":   false, // 6.4 gained Bearer + "username"
		"6.4.15":  false,
		"6.2.9":   true, // 6.2 -> body auth
		"6.0.28":  true, // 6.0 LTS -> body auth
		"6.0":     true,
		"5.4.0":   true, // 5.x -> body auth
		"5.0.0":   true,
		"":        false, // unknown -> default 7.x
		"garbage": false,
	}
	for v, want := range cases {
		if got := useBodyAuth(v); got != want {
			t.Errorf("useBodyAuth(%q) = %v, want %v", v, got, want)
		}
	}
}
