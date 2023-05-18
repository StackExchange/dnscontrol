package dnssort

type stubRecord struct {
	name         string
	dependencies []string
}

func (record stubRecord) GetType() ChangeType {
	return Change
}

func (record stubRecord) GetNameFQDN() string {
	return record.name
}

func (record stubRecord) GetFQDNDependencies() []string {
	return record.dependencies
}

func (record stubRecord) Equals(change SortableChange) bool {
	return record.name == change.GetNameFQDN()
}

func stubRecordsAsSortableRecords(records []stubRecord) []SortableChange {
	sortableRecords := make([]SortableChange, len(records))
	for iX := range records {
		sortableRecords[iX] = records[iX]
	}

	return sortableRecords
}
