package diff2

import (
	"github.com/StackExchange/dnscontrol/v3/pkg/dnssort"
)

func orderByDependencies(changes ChangeList) ChangeList {
	a := dnssort.SortUsingGraph(changes)

	return a
}
