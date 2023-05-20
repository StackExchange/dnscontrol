package dnssort

type stubRecord struct {
	NameFQDN     string
	Dependencies []Dependency
	Type         ChangeType
}

func (record stubRecord) GetType() ChangeType {
	return record.Type
}

func (record stubRecord) GetNameFQDN() string {
	return record.NameFQDN
}

func (record stubRecord) GetFQDNDependencies() []Dependency {
	return record.Dependencies
}

func (record stubRecord) Equals(change SortableChange) bool {
	return record.NameFQDN == change.GetNameFQDN()
}

func stubRecordsAsSortableRecords(records []stubRecord) []SortableChange {
	sortableRecords := make([]SortableChange, len(records))
	for iX := range records {
		sortableRecords[iX] = records[iX]
	}

	return sortableRecords
}
