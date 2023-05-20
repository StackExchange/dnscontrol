package dnssort

import "testing"

func Test_SortByDependencies(t *testing.T) {

	t.Run("Direct dependency",
		executeUniformSort(
			[]stubRecord{
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}}},
				{NameFQDN: "example.com", Dependencies: []Dependency{}},
			},
			[]string{
				"example.com",
				"www.example.com",
			},
			[]string{},
		),
	)

	t.Run("Already in correct order",
		executeUniformSort(
			[]stubRecord{
				{NameFQDN: "example.com", Dependencies: []Dependency{}},
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}}},
			},
			[]string{
				"example.com",
				"www.example.com",
			},
			[]string{},
		),
	)

	t.Run("Use wildcards",
		executeUniformSort(
			[]stubRecord{
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "a.test.example.com"}}},
				{NameFQDN: "*.test.example.com", Dependencies: []Dependency{}},
			},
			[]string{
				"*.test.example.com",
				"www.example.com",
			},
			[]string{},
		),
	)

	t.Run("Cyclic dependency added on the end",
		executeUniformSort(
			[]stubRecord{
				{NameFQDN: "a.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "b.example.com"}}},
				{NameFQDN: "b.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "a.example.com"}}},
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}}},
				{NameFQDN: "example.com", Dependencies: []Dependency{}},
			},
			[]string{
				"example.com",
				"www.example.com",
				"a.example.com",
				"b.example.com",
			},
			[]string{
				"a.example.com",
				"b.example.com",
			},
		),
	)

	t.Run("Self dependency added on the end",
		executeUniformSort(
			[]stubRecord{
				{NameFQDN: "a.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "a.example.com"}}},
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}}},
				{NameFQDN: "example.com", Dependencies: []Dependency{{}}},
			},
			[]string{
				"example.com",
				"www.example.com",
				"a.example.com",
			},
			[]string{
				"a.example.com",
			},
		),
	)

	t.Run("Ignores external dependency",
		executeUniformSort(
			[]stubRecord{
				{NameFQDN: "mail.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "mail.external.tld"}}},
			},
			[]string{
				"mail.example.com",
			},
			[]string{},
		),
	)

}

func getRecordsNames(records []SortableChange) []string {
	names := make([]string, len(records))
	for iX, record := range records {
		names[iX] = record.GetNameFQDN()
	}
	return names
}

func executeUniformSort(inputOrder []stubRecord, expectedOutputOrder []string, expectedUnresolved []string) func(*testing.T) {
	return func(t *testing.T) {
		inputSortableRecords := stubRecordsAsSortableRecords(inputOrder)
		t.Helper()
		result := SortByDependencies(inputSortableRecords)

		for iX := range expectedOutputOrder {
			if result.SortedRecords[iX].GetNameFQDN() != expectedOutputOrder[iX] {
				t.Errorf("Invalid order on index %d after sort. Expected order: %v. Got order: %v\n", iX, expectedOutputOrder, getRecordsNames(result.SortedRecords))
			}
		}
		if len(expectedOutputOrder) != len(result.SortedRecords) {
			t.Errorf("Missing records after sort. Expected order: %v. Got order: %v\n", expectedOutputOrder, getRecordsNames(result.SortedRecords))
		}

		for iX := range expectedUnresolved {
			if result.UnresolvedRecords[iX].GetNameFQDN() != expectedUnresolved[iX] {
				t.Errorf("Invalid unresolved records after sort. Expected: %v. Got: %v\n", expectedOutputOrder, getRecordsNames(result.UnresolvedRecords))
			}
		}
		if len(expectedUnresolved) != len(result.UnresolvedRecords) {
			t.Errorf("Missing unresolved records. Expected unresolved: %v. Got: %v\n", expectedUnresolved, getRecordsNames(result.UnresolvedRecords))
		}
	}
}
