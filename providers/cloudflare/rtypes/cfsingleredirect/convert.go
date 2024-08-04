package cfsingleredirect

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// TranscodePRtoSR takes a PAGE_RULE record, stores transcoded versions of the fields, and makes the record a CLOUDFLAREAPI_SINGLE_REDDIRECT.
func TranscodePRtoSR(rec *models.RecordConfig) error {
	rec.Type = SINGLEREDIRECT // This record is now a CLOUDFLAREAPI_SINGLE_REDIRECT

	// Extract the fields we're reading from:
	sr := rec.CloudflareRedirect
	code := sr.Code
	prWhen := sr.PRWhen
	prThen := sr.PRThen
	srName := sr.PRDisplay

	// Convert old-style patterns to new-style rules:
	srWhen, srThen, err := makeRuleFromPattern(prWhen, prThen)
	if err != nil {
		return err
	}

	// Fix the RecordConfig
	makeSingleRedirectFromConvert(rec,
		sr.PRPriority,
		prWhen, prThen,
		code,
		srName, srWhen, srThen)

	return nil
}

// makeRuleFromPattern compile old-style patterns and replacements into new-style rules and expressions.
func makeRuleFromPattern(pattern, replacement string) (string, string, error) {

	var srWhen, srThen string
	var err error

	var phost, ppath string
	origPattern := pattern
	pattern, phost, ppath, err = normalizeURL(pattern)
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

	// "pr" is Page Rule (old style)
	// "sr" is Static Rule (new style)
	// prWhen + prThen is the old-style matching pattern and replacement pattern.
	// srWhen + srThen is the new-style matching rule and replacement expression.

	if !strings.Contains(phost, `*`) && (ppath == `/` || ppath == "") {
		// https://i.sstatic.net/  (No Wildcards)
		srWhen = fmt.Sprintf(`http.host eq "%s" and http.request.uri.path eq "%s"`, phost, "/")

	} else if !strings.Contains(phost, `*`) && (ppath == `/*`) {
		// https://i.stack.imgur.com/*
		srWhen = fmt.Sprintf(`http.host eq "%s"`, phost)

	} else if !strings.Contains(phost, `*`) && !strings.Contains(ppath, "*") {
		// https://insights.stackoverflow.com/trends
		srWhen = fmt.Sprintf(`http.host eq "%s" and http.request.uri.path eq "%s"`, phost, ppath)

	} else if phost[0] == '*' && strings.Count(phost, `*`) == 1 && !strings.Contains(ppath, "*") {
		// *stackoverflow.careers/  (wildcard at beginning only)
		srWhen = fmt.Sprintf(`( http.host eq "%s" or ends_with(http.host, ".%s") ) and http.request.uri.path eq "%s"`, phost[1:], phost[1:], ppath)

	} else if phost[0] == '*' && strings.Count(phost, `*`) == 1 && ppath == "/*" {
		// *stackoverflow.careers/*  (wildcard at beginning and end)
		srWhen = fmt.Sprintf(`http.host eq "%s" or ends_with(http.host, ".%s")`, phost[1:], phost[1:])

	} else if strings.Contains(phost, `*`) && ppath == "/*" {
		// meta.*yodeya.com/* (wildcard in host)
		h := simpleGlobToRegex(phost)
		srWhen = fmt.Sprintf(`http.host matches r###"%s"###`, h)

	} else if !strings.Contains(phost, `*`) && strings.Count(ppath, `*`) == 1 && strings.HasSuffix(ppath, "*") {
		// domain.tld/.well-known* (wildcard in path)
		srWhen = fmt.Sprintf(`(starts_with(http.request.uri.path, "%s") and http.host eq "%s")`,
			ppath[0:len(ppath)-1],
			phost)

	}

	// replacement

	if !strings.Contains(replacement, `$`) {
		//  https://stackexchange.com/ (no substitutions)
		srThen = fmt.Sprintf(`concat("%s", "")`, replacement)

	} else if phost[0] == '*' && strings.Count(phost, `*`) == 1 && strings.Count(replacement, `$`) == 1 && len(rpath) > 3 && strings.HasSuffix(rpath, "/$2") {
		// *stackoverflowenterprise.com/* -> https://www.stackoverflowbusiness.com/enterprise/$2
		srThen = fmt.Sprintf(`concat("https://%s", "%s", http.request.uri.path)`,
			rhost,
			rpath[0:len(rpath)-3],
		)

	} else if phost[0] == '*' && strings.Count(phost, `*`) == 1 && strings.Count(replacement, `$`) == 1 && len(rpath) > 3 && strings.HasSuffix(rpath, "/$2") {
		// *stackoverflowenterprise.com/* -> https://www.stackoverflowbusiness.com/enterprise/$2
		srThen = fmt.Sprintf(`concat("https://%s", "%s", http.request.uri.path)`,
			rhost,
			rpath[0:len(rpath)-3],
		)

	} else if strings.Count(replacement, `$`) == 1 && rpath == `/$1` {
		// https://i.sstatic.net/$1 ($1 at end)
		srThen = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, rhost)

	} else if strings.Count(phost, `*`) == 1 && strings.Count(ppath, `*`) == 1 &&
		strings.Count(replacement, `$`) == 1 && strings.HasSuffix(rpath, `/$2`) {
		// https://careers.stackoverflow.com/$2
		srThen = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, rhost)

	} else if strings.Count(replacement, `$`) == 1 && strings.HasSuffix(replacement, `$1`) {
		// https://social.domain.tld/.well-known$1
		srThen = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, rhost)

	} else if strings.Count(replacement, `$`) == 1 && strings.HasSuffix(replacement, `$1`) {
		// https://social.domain.tld/.well-known$1
		srThen = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, rhost)

	}

	// Not implemented

	if srWhen == "" {
		return "", "", fmt.Errorf("conversion not implemented for pattern: %s", origPattern)
	}
	if srThen == "" {
		return "", "", fmt.Errorf("conversion not implemented for replacement: %s", origReplacement)
	}

	return srWhen, srThen, nil
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
