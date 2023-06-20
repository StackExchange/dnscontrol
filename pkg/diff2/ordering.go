package diff2

import (
	"log"

	"github.com/StackExchange/dnscontrol/v4/pkg/dnsgraph"
	"github.com/StackExchange/dnscontrol/v4/pkg/dnssort"
)

func orderByDependencies(changes ChangeList) ChangeList {
	a := dnssort.SortUsingGraph(changes)

	if len(a.UnresolvedRecords) > 0 {
		log.Printf("Found unresolved records %v.\n"+
			"This can indicate a circular dependency, please ensure all targets from given records exist and no circular dependencies exist in the changeset."+
			"These unresolved records are still added as changes and pushed to the provider, but will cause issues if and when the provider checks the changes\n",
			dnsgraph.GetRecordsNamesForGraphables(a.UnresolvedRecords),
		)
	}

	return a.SortedRecords
}
