package dnssort

import "log"

type directedSortState struct {
	graph                *DNSGraph
	sortedRecords        []SortableChange
	unresolvedRecords    []SortableChange
	hasResolvedLastRound bool
}

func createDirectedSortState(records []SortableChange) directedSortState {
	graph := CreateGraph(records)

	return directedSortState{
		graph:                graph,
		unresolvedRecords:    []SortableChange{},
		sortedRecords:        []SortableChange{},
		hasResolvedLastRound: false,
	}
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

			if !nodeIsResolved(node) {
				continue
			}

			sortState.hasResolvedLastRound = true
			sortState.sortedRecords = append(sortState.sortedRecords, node.change)
			sortState.graph.removeNode(node)
		}

		if sortState.hasStalled() {
			break
		}
	}

	if len(sortState.graph.all) > 0 {
		log.Printf("The DNS changes appear to have unresolved dependencies like %s\n", sortState.graph.all[0].change.GetNameFQDN())
		for _, unresolved := range sortState.graph.all {
			sortState.sortedRecords = append(sortState.sortedRecords, unresolved.change)
			sortState.unresolvedRecords = append(sortState.unresolvedRecords, unresolved.change)
		}
	}

	return SortResult{
		SortedRecords:     sortState.sortedRecords,
		UnresolvedRecords: sortState.unresolvedRecords,
	}
}

func nodeIsResolved(node *dnsGraphNode) bool {
	return node.change.GetType() == DeletionChange && len(node.incoming) == 0 || node.change.GetType() == AdditionChange && len(node.outgoing) == 0
}
