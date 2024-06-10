package cloudflare

import (
	"fmt"
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

func makeRuleFromPattern(pattern, replacement string, temporary bool) (string, string, error) {
	// TODO: Change this function to do something useful with the replacement string.

	// TODO: These are just to get rid of the warning that the variables are unused.
	_ = replacement
	_ = temporary

	var matcher, expr string

	// TODO: replace with a real conversion.
	if pattern == `example.com/` {
		matcher = fmt.Sprintf(`http.host eq "%s" and http.request.uri.path eq "%s"`, "example.com", "/")
		expr = fmt.Sprintf(`concat("https://%s", http.request.uri.path)`, "example.com")
		return matcher, expr, nil
	}

	return "", "", fmt.Errorf("unimplemented: %s", pattern)
}
