package dnssort

import (
	"log"
)

type uniformSortState struct {
	availableNames       *DomainTree[interface{}]
	resolvedNames        *DomainTree[interface{}]
	sortedRecords        []SortableChange
	unresolvedRecords    []SortableChange
	workingSet           []unresolvedRecord
	nextWorkingSet       []unresolvedRecord
	hasResolvedLastRound bool
}

type unresolvedRecord struct {
	unresolvedDependencies nameMap
	record                 SortableChange
}

func createUniformSortState(records []SortableChange) uniformSortState {
	sortState := uniformSortState{
		availableNames:       CreateTree[interface{}](),
		resolvedNames:        CreateTree[interface{}](),
		sortedRecords:        make([]SortableChange, 0),
		unresolvedRecords:    make([]SortableChange, 0),
		workingSet:           nil,
		nextWorkingSet:       nil,
		hasResolvedLastRound: false,
	}

	for _, record := range records {
		sortState.availableNames.Set(record.GetNameFQDN(), struct{}{})
	}

	for _, record := range records {
		sortState.workingSet = append(sortState.workingSet, sortState.createUnresolvedRecordFor(record))
	}

	return sortState
}

func (sortState *uniformSortState) addAsResolved(unresolved unresolvedRecord) {
	sortState.hasResolvedLastRound = true
	sortState.sortedRecords = append(sortState.sortedRecords, unresolved.record)
	sortState.resolvedNames.Set(unresolved.record.GetNameFQDN(), struct{}{})
}

func (sortState *uniformSortState) createUnresolvedRecordFor(record SortableChange) unresolvedRecord {
	unresolvedDependencies := nameMap{}

	for _, fqdn := range record.GetFQDNDependencies() {
		if sortState.availableNames.Has(fqdn.NameFQDN) {
			unresolvedDependencies.add(fqdn.NameFQDN)
		}
	}

	return unresolvedRecord{
		unresolvedDependencies: unresolvedDependencies,
		record:                 record,
	}
}

func (sortState *uniformSortState) swapWorkingSets() {
	sortState.workingSet = sortState.nextWorkingSet
	sortState.nextWorkingSet = nil
}

func (sortState *uniformSortState) removeResolvedNames(unresolved unresolvedRecord) {
	for dependency := range unresolved.unresolvedDependencies {
		if sortState.resolvedNames.Has(dependency) {
			unresolved.unresolvedDependencies.remove(dependency)
		}
	}
}

func (sortState *uniformSortState) addForNextRound(unresolved unresolvedRecord) {
	sortState.nextWorkingSet = append(sortState.nextWorkingSet, unresolved)
}

func (sortState *uniformSortState) hasWork() bool {
	return len(sortState.workingSet) > 0
}

func (sortState *uniformSortState) hasStalled() bool {
	return !sortState.hasResolvedLastRound
}

type SortResult struct {
	SortedRecords     []SortableChange
	UnresolvedRecords []SortableChange
}

func SortByDependencies(records []SortableChange) SortResult {
	sortState := createUniformSortState(records)

	for sortState.hasWork() {
		sortState.hasResolvedLastRound = false

		for _, unresolved := range sortState.workingSet {
			sortState.removeResolvedNames(unresolved)

			if !unresolved.unresolvedDependencies.empty() {
				sortState.addForNextRound(unresolved)
			} else {
				sortState.addAsResolved(unresolved)
			}
		}

		if sortState.hasStalled() {
			log.Printf("The DNS changes appear to have unresolved dependencies like %s\n", sortState.workingSet[0].unresolvedDependencies)
			for _, unresolved := range sortState.workingSet {
				sortState.sortedRecords = append(sortState.sortedRecords, unresolved.record)
				sortState.unresolvedRecords = append(sortState.unresolvedRecords, unresolved.record)
			}
			break
		}

		sortState.swapWorkingSets()
	}

	return SortResult{
		SortedRecords:     sortState.sortedRecords,
		UnresolvedRecords: sortState.unresolvedRecords,
	}
}
