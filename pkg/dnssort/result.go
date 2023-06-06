package dnssort

import "github.com/StackExchange/dnscontrol/v4/pkg/dnsgraph"

type SortResult[T dnsgraph.Graphable] struct {
	SortedRecords     []T
	UnresolvedRecords []T
}
