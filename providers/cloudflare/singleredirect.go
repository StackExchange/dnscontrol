package cloudflare

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func newCfsrFromUserInput(target string, code int, priority int) (*models.CloudflareSingleRedirectConfig, error) {
	// target: matcher,replacement,priority,code
	// target: cable.slackoverflow.com/*,https://change.cnn.com/$1,1,302

	r := &models.CloudflareSingleRedirectConfig{}

	// Break apart the 4-part string and store into the individual fields:
	parts := strings.Split(target, ",")
	//printer.Printf("DEBUG: cfsrFromOldStyle: parts=%v\n", parts)
	r.PRDisplay = fmt.Sprintf("%s,%d,%03d", target, priority, code)
	r.PRMatcher = parts[0]
	r.PRReplacement = parts[1]
	r.PRPriority = priority
	r.Code = code

	// Convert old-style to new-style:
	if err := addNewStyleFields(r); err != nil {
		return nil, err
	}
	return r, nil
}

func newCfsrFromAPIData(sm, sr string, code int) *models.CloudflareSingleRedirectConfig {
	r := &models.CloudflareSingleRedirectConfig{
		PRMatcher:     "UNKNOWABLE",
		PRReplacement: "UNKNOWABLE",
		//PRPriority:    0,
		Code:          code,
		SRMatcher:     sm,
		SRReplacement: sr,
	}
	return r
}

// addNewStyleFields takes a PAGE_RULE-style target and populates the CFSRC.
func addNewStyleFields(sr *models.CloudflareSingleRedirectConfig) error {

	// Extract the fields we're reading from:
	prMatcher := sr.PRMatcher
	prReplacement := sr.PRReplacement
	code := sr.Code

	// Convert old-style patterns to new-style rules:
	srMatcher, srReplacement, err := makeRuleFromPattern(prMatcher, prReplacement, code != 301)
	if err != nil {
		return err
	}
	display := fmt.Sprintf(`%s,%s,%d,%03d matcher=%q replacement=%q`,
		prMatcher, prReplacement,
		sr.PRPriority, code,
		srMatcher, srReplacement,
	)

	// Store the results in the fields we're writing to:
	sr.SRMatcher = srMatcher
	sr.SRReplacement = srReplacement
	sr.SRDisplay = display

	return nil
}

// makeRuleFromPattern compile old-style patterns and replacements into new-style rules and expressions.
func makeRuleFromPattern(pattern, replacement string, temporary bool) (string, string, error) {

	_ = temporary // Prevents error due to this variable not (yet) being used

	var matcher, expr string
	var err error

	var host, path string
	origPattern := pattern
	pattern, host, path, err = normalizeURL(pattern)
	_ = pattern
	if err != nil {
		return "", "", err
	}
	var rhost, rpath string
	origReplacement := replacement
	replacement, rhost, rpath, err = normalizeURL(replacement)
	_ = rpath
	if err != nil {
		return "", "", err
	}

	// TODO(tlim): This could be a lot faster by not repeating itself so much.
	// However I want to get it working before it is optimized.

	// pattern -> matcher

	if !strings.Contains(host, `*`) && (path == `/` || path == "") {
		// https://i.sstatic.net/  (No Wildcards)
		matcher = fmt.Sprintf(`http.host eq "%s" and http.request.uri.path eq "%s"`, host, "/")

	} else if !strings.Contains(host, `*`) && (path == `/*`) {
		// https://i.stack.imgur.com/*
		matcher = fmt.Sprintf(`http.host eq "%s"`, host)

	} else if !strings.Contains(host, `*`) && !strings.Contains(path, "*") {
		// https://insights.stackoverflow.com/trends
		matcher = fmt.Sprintf(`http.host eq "%s" and http.request.uri.path eq "%s"`, host, path)

	} else if host[0] == '*' && strings.Count(host, `*`) == 1 && !strings.Contains(path, "*") {
		// *stackoverflow.careers/  (wildcard at beginning only)
		matcher = fmt.Sprintf(`( http.host eq "%s" or ends_with(http.host, ".%s") ) and http.request.uri.path eq "%s"`, host[1:], host[1:], path)

	} else if host[0] == '*' && strings.Count(host, `*`) == 1 && path == "/*" {
		// *stackoverflow.careers/*  (wildcard at beginning and end)
		matcher = fmt.Sprintf(`http.host eq "%s" or ends_with(http.host, ".%s")`, host[1:], host[1:])

	} else if strings.Contains(host, `*`) && path == "/*" {
		// meta.*yodeya.com/* (wildcard in host)
		h := simpleGlobToRegex(host)
		matcher = fmt.Sprintf(`http.host matches r###"%s"###`, h)

	} else if !strings.Contains(host, `*`) && strings.Count(path, `*`) == 1 && strings.HasSuffix(path, "*") {
		// domain.tld/.well-known* (wildcard in path)
		matcher = fmt.Sprintf(`(starts_with(http.request.uri.path, "%s") and http.host eq "%s")`,
			path[0:len(path)-1],
			host)

	}

	// replacement

	if !strings.Contains(replacement, `$`) {
		//  https://stackexchange.com/ (no substitutions)
		expr = fmt.Sprintf(`concat("%s", "")`, replacement)

	} else if host[0] == '*' && strings.Count(host, `*`) == 1 && strings.Count(replacement, `$`) == 1 && len(rpath) > 3 && strings.HasSuffix(rpath, "/$2") {
		// *stackoverflowenterprise.com/* -> https://www.stackoverflowbusiness.com/enterprise/$2
		expr = fmt.Sprintf(`concat("https://%s", "%s", http.request.uri.path)`,
			rhost,
			rpath[0:len(rpath)-3],
		)

	} else if strings.Count(replacement, `$`) == 1 && rpath == `/$1` {
		// https://i.sstatic.net/$1 ($1 at end)
		expr = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, rhost)

	} else if strings.Count(host, `*`) == 1 && strings.Count(path, `*`) == 1 &&
		strings.Count(replacement, `$`) == 1 && strings.HasSuffix(rpath, `/$2`) {
		// https://careers.stackoverflow.com/$2
		expr = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, rhost)

	} else if strings.Count(replacement, `$`) == 1 && strings.HasSuffix(replacement, `$1`) {
		// https://social.domain.tld/.well-known$1
		expr = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, rhost)

	}

	// Not implemented

	if matcher == "" {
		return "", "", fmt.Errorf("conversion not implemented for pattern: %s", origPattern)
	}
	if expr == "" {
		return "", "", fmt.Errorf("conversion not implemented for replacement: %s", origReplacement)
	}

	return matcher, expr, nil
}

// normalizeURL turns foo.com into https://foo.com and replaces HTTP with HTTPS.
// It also returns an error if there is a port specified (like :8080)
func normalizeURL(s string) (string, string, string, error) {
	orig := s
	if strings.HasPrefix(s, `http://`) {
		s = "https://" + s[7:]
	} else if !strings.HasPrefix(s, `https://`) {
		s = `https://` + s
	}

	// Make sure it parses.
	u, err := url.Parse(s)
	if err != nil {
		return "", "", "", err
	}

	// Make sure it doesn't have a port (https://example.com:8080)
	_, port, _ := net.SplitHostPort(u.Host)
	if port != "" {
		return "", "", "", fmt.Errorf("unimplemented port: %q", orig)
	}

	return s, u.Host, u.Path, nil
}

// simpleGlobToRegex translates very simple Glob patterns into regexp-compatible expressions.
// It only handles `.` and `*` currently.  See singleredirect_test.go for supported patterns.
func simpleGlobToRegex(g string) string {

	if g == "" {
		return `.*`
	}

	if !strings.HasSuffix(g, "*") {
		g = g + `$`
	}
	if !strings.HasPrefix(g, "*") {
		g = `^` + g
	}

	g = strings.ReplaceAll(g, `.`, `\.`)
	g = strings.ReplaceAll(g, `*`, `.*`)
	return g
}
