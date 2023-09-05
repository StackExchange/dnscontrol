package testutils

import "github.com/StackExchange/dnscontrol/v4/pkg/dnsgraph"

type StubRecord struct {
	NameFQDN     string
	Dependencies []dnsgraph.Dependency
	Type         dnsgraph.NodeType
}

func (record StubRecord) GetType() dnsgraph.NodeType {
	return record.Type
}

func (record StubRecord) GetName() string {
	return record.NameFQDN
}

func (record StubRecord) GetDependencies() []dnsgraph.Dependency {
	return record.Dependencies
}

func StubRecordsAsGraphable(records []StubRecord) []dnsgraph.Graphable {
	sortableRecords := make([]dnsgraph.Graphable, len(records))

	for iX := range records {
		sortableRecords[iX] = records[iX]
	}

	return sortableRecords
}
