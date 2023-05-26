package dnssort

type unresolvedRecord[T SortableChange] struct {
	unresolvedDependencies nameMap
	record                 T
}

type SortResult[T SortableChange] struct {
	SortedRecords     []T
	UnresolvedRecords []T
}
