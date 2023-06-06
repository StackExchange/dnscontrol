package diff2

import (
	"log"

	"github.com/StackExchange/dnscontrol/v4/pkg/dnsgraph"
	"github.com/StackExchange/dnscontrol/v4/pkg/dnssort"
)

func orderByDependencies(changes ChangeList) ChangeList {
	a := dnssort.SortUsingGraph(changes)

	if len(a.UnresolvedRecords) > 0 {
		log.Printf("Found unresolved records %v\n", dnsgraph.GetRecordsNamesForGraphables(a.UnresolvedRecords))
	}

	return a.SortedRecords
}
