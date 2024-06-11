package cloudflare

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// generateSingleRedirectRule takes a PAGE_RULE-style target and returns strings
// to use with Single Redirect.
// target format is: pattern,action,index,code
//
//	pattern: The glob-like pattern
//	action: The replacement pattern (with $1, $2, etc substitutions)
//	index: The index in the list of rules. This is ignored.
//	code: 301 or 302
func generateSingleRedirectRule(target string) (string, string, string, error) {
	// FIXME(tlim): Instead of returning so many strings, this should probably return a struct.
	//      Possibly cloudflare.RulesetRule or cloudflare.RulesetRuleActionParameters ?

	parts := strings.Split(target, ",")
	constraint := parts[0]
	action := parts[1]
	code, _ := strconv.Atoi(parts[3])

	pattern, replacement, err := makeRuleFromPattern(constraint, action, code != 301)
	target = fmt.Sprintf("%03d,%q,%q matcher=%q expr=%s", code, constraint, action, pattern, replacement)
	return pattern, replacement, target, err
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
	}

	// replacement

	if !strings.Contains(replacement, `$`) {
		//  https://stackexchange.com/ (no substitutions)
		expr = fmt.Sprintf(`"%s"`, replacement)

	} else if strings.Count(replacement, `$`) == 1 && rpath == `/$1` {
		// https://i.sstatic.net/$1 ($1 at end)
		expr = fmt.Sprintf(`concat("https://%s/", http.request.uri.path)`, rhost)

	} else if strings.Count(host, `*`) == 1 && strings.Count(path, `*`) == 1 &&
		strings.Count(replacement, `$`) == 1 && rpath == `/$2` {
		// https://careers.stackoverflow.com/$2
		expr = fmt.Sprintf(`concat("https://%s/", http.request.uri.path)`, rhost)

	}

	// Not implemented

	if matcher == "" {
		return "", "", fmt.Errorf("conversion not implemented for pattern: %s", origPattern)
	}
	if expr == "" {
		return "", "", fmt.Errorf("conversion not implemented for replacemennt: %s", origReplacement)
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
