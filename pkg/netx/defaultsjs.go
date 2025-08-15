package netx

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SOCKS5FromDefaultsJS scans a local "defaults.js" (in CWD) for DNSCONTROL_SOCKS5.
// It supports either of the following syntaxes:
//   DNSCONTROL_SOCKS5 = "socks5://host:1080"
//   CLI_DEFAULTS({ "DNSCONTROL_SOCKS5": "socks5://host:1080" })
// Returns "" if file doesn't exist or no match is found.
func SOCKS5FromDefaultsJS() string {
	wd, _ := os.Getwd()
	if wd == "" {
		return ""
	}
	path := filepath.Join(wd, "defaults.js")
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	txt := string(b)

	// 1) DNSCONTROL_SOCKS5 = "...."
	reEq := regexp.MustCompile(`(?m)DNSCONTROL_SOCKS5\s*=\s*"(.*?)"`)
	if m := reEq.FindStringSubmatch(txt); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}

	// 2) CLI_DEFAULTS({ "DNSCONTROL_SOCKS5": "...." })
	reJSON := regexp.MustCompile(`DNSCONTROL_SOCKS5"\s*:\s*"(.*?)"`)
	if m := reJSON.FindStringSubmatch(txt); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}

	return ""
}
