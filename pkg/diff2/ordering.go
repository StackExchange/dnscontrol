package diff2

import (
	"log"

	"github.com/StackExchange/dnscontrol/v3/pkg/dnssort"
)

func orderByDependencies(changes ChangeList) ChangeList {
	a := dnssort.SortUsingGraph(changes)

	if len(a.UnresolvedRecords) > 0 {
		log.Fatalf("Found unresolved records %v\n", dnssort.GetRecordsNamesForChanges(a.UnresolvedRecords))
	}

	return a.SortedRecords
}
