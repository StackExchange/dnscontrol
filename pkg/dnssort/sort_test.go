package dnssort

import "testing"

func Test_SortByDependencies(t *testing.T) {

	t.Run("Direct dependency",
		executeSort(
			[]stubRecord{
				{name: "www.example.com", dependencies: []string{"example.com"}},
				{name: "example.com", dependencies: []string{}},
			},
			[]string{
				"example.com",
				"www.example.com",
			},
			[]string{},
		),
	)

	t.Run("Already in correct order",
		executeSort(
			[]stubRecord{
				{name: "example.com", dependencies: []string{}},
				{name: "www.example.com", dependencies: []string{"example.com"}},
			},
			[]string{
				"example.com",
				"www.example.com",
			},
			[]string{},
		),
	)

	t.Run("Use wildcards",
		executeSort(
			[]stubRecord{
				{name: "www.example.com", dependencies: []string{"a.test.example.com"}},
				{name: "*.test.example.com", dependencies: []string{}},
			},
			[]string{
				"*.test.example.com",
				"www.example.com",
			},
			[]string{},
		),
	)

	t.Run("Cyclic dependency added on the end",
		executeSort(
			[]stubRecord{
				{name: "a.example.com", dependencies: []string{"b.example.com"}},
				{name: "b.example.com", dependencies: []string{"a.example.com"}},
				{name: "www.example.com", dependencies: []string{"example.com"}},
				{name: "example.com", dependencies: []string{}},
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
		executeSort(
			[]stubRecord{
				{name: "a.example.com", dependencies: []string{"a.example.com"}},
				{name: "www.example.com", dependencies: []string{"example.com"}},
				{name: "example.com", dependencies: []string{}},
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
		executeSort(
			[]stubRecord{
				{name: "mail.example.com", dependencies: []string{"mail.external.tld"}},
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

func executeSort(inputOrder []stubRecord, expectedOutputOrder []string, expectedUnresolved []string) func(*testing.T) {
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
