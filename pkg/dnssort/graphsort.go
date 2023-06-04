package dnssort

type directedSortState[T SortableChange] struct {
	graph                *DNSGraph[T]
	sortedRecords        []T
	unresolvedRecords    []T
	hasResolvedLastRound bool
}

func createDirectedSortState[T SortableChange](records []T) directedSortState[T] {
	changes, reportChanges := splitRecordsByType(records)

	graph := CreateGraph(changes)

	return directedSortState[T]{
		graph:                graph,
		unresolvedRecords:    []T{},
		sortedRecords:        reportChanges,
		hasResolvedLastRound: false,
	}
}

func splitRecordsByType[T SortableChange](records []T) ([]T, []T) {
	var changes []T
	var reports []T

	for _, record := range records {
		switch record.GetType() {
		case Report:
			reports = append(reports, record)
		case Change:
			changes = append(changes, record)
		}
	}

	return changes, reports
}

func (sortState *directedSortState[T]) hasWork() bool {
	return len(sortState.graph.all) > 0
}

func (sortState *directedSortState[T]) hasStalled() bool {
	return !sortState.hasResolvedLastRound
}

func SortUsingGraph[T SortableChange](records []T) SortResult[T] {
	sortState := createDirectedSortState(records)

	for sortState.hasWork() {

		for _, node := range sortState.graph.all {
			sortState.hasResolvedLastRound = false

			if node.hasUnmetDependencies() {
				continue
			}

			sortState.hasResolvedLastRound = true
			sortState.addSortedRecord(node.Change)
			sortState.graph.removeNode(node)
		}

		if sortState.hasStalled() {
			break
		}
	}

	sortState.finalize()

	return SortResult[T]{
		SortedRecords:     sortState.sortedRecords,
		UnresolvedRecords: sortState.unresolvedRecords,
	}
}

func (sortState *directedSortState[T]) addSortedRecord(node T) {
	sortState.sortedRecords = append(sortState.sortedRecords, node)
}

func (sortState *directedSortState[T]) finalize() {
	// Add all of the changes remaining in the graph as unresolved and add them at the end of the sorted result to at least include everything
	if len(sortState.graph.all) > 0 {
		for _, unresolved := range sortState.graph.all {
			sortState.addSortedRecord(unresolved.Change)
			sortState.unresolvedRecords = append(sortState.unresolvedRecords, unresolved.Change)
		}
	}
}

func (node *dnsGraphNode[T]) hasUnmetDependencies() bool {
	for _, edge := range node.Edges {
		if edge.Dependency.Type == BackwardDependency && edge.Direction == IncomingEdge {
			return true
		}

		if edge.Dependency.Type == ForwardDependency && edge.Direction == OutgoingEdge {
			return true
		}
	}

	return false
}
