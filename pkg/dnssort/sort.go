package dnssort

import (
	"log"
)

type ChangeType uint8

const (
	Add ChangeType = iota
	Delete
	Change
)

type SortableChange interface {
	GetType() ChangeType
	GetNameFQDN() string
	GetFQDNDependencies() []string
	Equals(change SortableChange) bool
}

type SortableRecords []SortableChange

type unresolvedRecord struct {
	unresolvedDependencies nameMap
	record                 SortableChange
}

type sortState struct {
	availableNames       *DomainTree[interface{}]
	resolvedNames        *DomainTree[interface{}]
	sortedRecords        []SortableChange
	unresolvedRecords    []SortableChange
	workingSet           []unresolvedRecord
	nextWorkingSet       []unresolvedRecord
	hasResolvedLastRound bool
}

func createSortState(records []SortableChange) sortState {
	sortState := sortState{
		availableNames:       CreateTree[interface{}](),
		resolvedNames:        CreateTree[interface{}](),
		sortedRecords:        make([]SortableChange, 0),
		unresolvedRecords:    make([]SortableChange, 0),
		workingSet:           nil,
		nextWorkingSet:       nil,
		hasResolvedLastRound: false,
	}

	for _, record := range records {
		sortState.availableNames.Add(record.GetNameFQDN(), struct{}{})
	}

	for _, record := range records {
		sortState.workingSet = append(sortState.workingSet, sortState.createUnresolvedRecordFor(record))
	}

	return sortState
}

func (sortState *sortState) addAsResolved(unresolved unresolvedRecord) {
	sortState.hasResolvedLastRound = true
	sortState.sortedRecords = append(sortState.sortedRecords, unresolved.record)
	sortState.resolvedNames.Add(unresolved.record.GetNameFQDN(), struct{}{})
}

func (sortState *sortState) createUnresolvedRecordFor(record SortableChange) unresolvedRecord {
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
	SortedRecords     []SortableChange
	UnresolvedRecords []SortableChange
}

func SortByDependencies(records []SortableChange) DependencySortResult {
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
