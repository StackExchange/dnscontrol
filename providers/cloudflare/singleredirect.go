package cloudflare

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func generateSingleRedirectRule(target string) (string, string, string, error) {
	// TODO: The list of returned items is growing long.  Maybe return a struct instead?
	//      Either cloudflare.RulesetRule or maybe cloudflare.RulesetRuleActionParameters

	parts := strings.Split(target, ",")
	constraint := parts[0]
	action := parts[1]
	code, _ := strconv.Atoi(parts[3])

	pattern, replacement, err := makeRuleFromPattern(constraint, action, code != 301)
	target = fmt.Sprintf("%03d,%q,%q matcher=%q expr=%s", code, constraint, action, pattern, replacement)
	return pattern, replacement, target, err
}

// normalizeURL turns foo.com into https://foo.com and replaces HTTP with HTTPS.
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

func makeRuleFromPattern(pattern, replacement string, temporary bool) (string, string, error) {
	// TODO: Change this function to do something useful with the replacement string.

	// TODO: These are just to get rid of the warning that the variables are unused.
	//_ = replacement
	_ = temporary

	var matcher, expr string
	var err error

	// TODO: replace with a real conversion.

	var host, path string
	origPattern := pattern
	pattern, host, path, err = normalizeURL(pattern)
	_ = pattern
	if err != nil {
		return "", "", err
	}
	var rhost, rpath string
	replacement, rhost, rpath, err = normalizeURL(replacement)
	_ = rpath
	if err != nil {
		return "", "", err
	}

	//  https://i.sstatic.net/,https://stackexchange.com/
	if !strings.Contains(host, `*`) && path == `/` && !strings.Contains(replacement, `$`) {
		matcher = fmt.Sprintf(`http.host eq "%s" and http.request.uri.path eq "%s"`, host, "/")
		expr = fmt.Sprintf(`"%s"`, replacement)
		return matcher, expr, nil
	}

	// https://i.stack.imgur.com/*,https://i.sstatic.net/$1
	if !strings.Contains(host, `*`) && path == `/*` && !strings.Contains(replacement, `$1`) {
		matcher = fmt.Sprintf(`http.host eq "%s" and http.request.uri.path eq "%s"`, host, "/*")
		expr = fmt.Sprintf(`concat("%s/", http.request.uri.path)`, rhost)
		return matcher, expr, nil
	}

	if matcher == "" {
		return "", "", fmt.Errorf("conversion not implemented for: %s", origPattern)
	}
	return matcher, expr, nil
}