package models

import "fmt"

var dotwarned = map[string]bool{}

func WarnNameserverDot(p, w string) {
	if dotwarned[p] {
		return
	}
	fmt.Printf("Warning: provider %s could be improved. See https://github.com/StackExchange/dnscontrol/issues/491 (%s)\n", p, w)
	dotwarned[p] = true
}
