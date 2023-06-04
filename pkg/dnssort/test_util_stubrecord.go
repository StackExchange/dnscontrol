package dnssort

type stubRecord struct {
	NameFQDN     string
	Dependencies []Dependency
	Type         ChangeType
}

func (record stubRecord) GetType() ChangeType {
	return record.Type
}

func (record stubRecord) GetName() string {
	return record.NameFQDN
}

func (record stubRecord) GetDependencies() []Dependency {
	return record.Dependencies
}

func stubRecordsAsSortableRecords(records []stubRecord) []SortableChange {
	sortableRecords := make([]SortableChange, len(records))

	for iX := range records {
		sortableRecords[iX] = records[iX]
	}

	return sortableRecords
}
