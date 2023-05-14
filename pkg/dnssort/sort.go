package dnssort

import (
	"log"
)

type SortableRecord interface {
	GetNameFQDN() string
	GetFQDNDependencies() []string
}

type SortableRecords []SortableRecord

type unresolvedRecord struct {
	unresolvedDependencies nameMap
	record                 SortableRecord
}

type sortState struct {
	availableNames       *DomainTree
	resolvedNames        *DomainTree
	sortedRecords        []SortableRecord
	unresolvedRecords    []SortableRecord
	workingSet           []unresolvedRecord
	nextWorkingSet       []unresolvedRecord
	hasResolvedLastRound bool
}

func createSortState(records []SortableRecord) sortState {
	sortState := sortState{
		availableNames:       CreateTree(),
		resolvedNames:        CreateTree(),
		sortedRecords:        make([]SortableRecord, 0),
		unresolvedRecords:    make([]SortableRecord, 0),
		workingSet:           nil,
		nextWorkingSet:       nil,
		hasResolvedLastRound: false,
	}

	for _, record := range records {
		sortState.availableNames.Add(record.GetNameFQDN())
	}

	for _, record := range records {
		sortState.workingSet = append(sortState.workingSet, sortState.createUnresolvedRecordFor(record))
	}

	return sortState
}

func (sortState *sortState) addAsResolved(unresolved unresolvedRecord) {
	sortState.hasResolvedLastRound = true
	sortState.sortedRecords = append(sortState.sortedRecords, unresolved.record)
	sortState.resolvedNames.Add(unresolved.record.GetNameFQDN())
}

func (sortState *sortState) createUnresolvedRecordFor(record SortableRecord) unresolvedRecord {
	unresolvedDependencies := nameMap{}

	for _, fqdn := range record.GetFQDNDependencies() {
		if sortState.availableNames.Has(fqdn) {
			unresolvedDependencies.add(fqdn)
		}
	}

	return unresolvedRecord{
		unresolvedDependencies: unresolvedDependencies,
		record:                 record,
	}
}

func (sortState *sortState) swapWorkingSets() {
	sortState.workingSet = sortState.nextWorkingSet
	sortState.nextWorkingSet = nil
}

func (sortState *sortState) removeResolvedNames(unresolved unresolvedRecord) {
	for dependency := range unresolved.unresolvedDependencies {
		if sortState.resolvedNames.Has(dependency) {
			unresolved.unresolvedDependencies.remove(dependency)
		}
	}
}

func (sortState *sortState) addForNextRound(unresolved unresolvedRecord) {
	sortState.nextWorkingSet = append(sortState.nextWorkingSet, unresolved)
}

func (sortState *sortState) hasWork() bool {
	return len(sortState.workingSet) > 0
}

func (sortState *sortState) hasStalled() bool {
	return !sortState.hasResolvedLastRound
}

type DependencySortResult struct {
	SortedRecords     []SortableRecord
	UnresolvedRecords []SortableRecord
}

func SortByDependencies(records []SortableRecord) DependencySortResult {
	sortState := createSortState(records)

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

	return DependencySortResult{
		SortedRecords:     sortState.sortedRecords,
		UnresolvedRecords: sortState.unresolvedRecords,
	}
}
