package dnssort

import "log"

type directedSortState struct {
	graph                *DNSGraph
	sortedRecords        []SortableChange
	unresolvedRecords    []SortableChange
	hasResolvedLastRound bool
}

func createDirectedSortState(records []SortableChange) directedSortState {
	changes, reportChanges := splitRecordsByType(records)

	graph := CreateGraph(changes)

	return directedSortState{
		graph:                graph,
		unresolvedRecords:    []SortableChange{},
		sortedRecords:        reportChanges,
		hasResolvedLastRound: false,
	}
}

func splitRecordsByType(records []SortableChange) ([]SortableChange, []SortableChange) {
	var changes []SortableChange
	var reports []SortableChange

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

func (sortState *directedSortState) hasWork() bool {
	return len(sortState.graph.all) > 0
}

func (sortState *directedSortState) hasStalled() bool {
	return !sortState.hasResolvedLastRound
}

func SortUsingGraph(records []SortableChange) SortResult {
	sortState := createDirectedSortState(records)

	for sortState.hasWork() {

		for _, node := range sortState.graph.all {
			sortState.hasResolvedLastRound = false

			if node.hasUnmetDependencies() {
				continue
			}

			sortState.hasResolvedLastRound = true
			sortState.sortedRecords = append(sortState.sortedRecords, node.Change)
			sortState.graph.removeNode(node)
		}

		if sortState.hasStalled() {
			break
		}
	}

	if len(sortState.graph.all) > 0 {
		log.Printf("The DNS changes appear to have unresolved dependencies like %s\n", sortState.graph.all[0].Change.GetNameFQDN())
		for _, unresolved := range sortState.graph.all {
			sortState.sortedRecords = append(sortState.sortedRecords, unresolved.Change)
			sortState.unresolvedRecords = append(sortState.unresolvedRecords, unresolved.Change)
		}
	}

	return SortResult{
		SortedRecords:     sortState.sortedRecords,
		UnresolvedRecords: sortState.unresolvedRecords,
	}
}

func (node *dnsGraphNode) hasUnmetDependencies() bool {
	for _, edge := range node.Edges {
		if edge.Dependency.Type == OldDependency && edge.Direction == IncomingEdge {
			return true
		}
		if edge.Dependency.Type == NewDependency && edge.Direction == OutgoingEdge {
			return true
		}
	}

	return false
}
