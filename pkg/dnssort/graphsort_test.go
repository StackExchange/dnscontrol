package dnssort

import "testing"

func Test_graphsort(t *testing.T) {

	t.Run("Direct dependency",
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}}},
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
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "example.com", Dependencies: []Dependency{}},
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}}},
			},
			[]string{
				"example.com",
				"www.example.com",
			},
			[]string{},
		),
	)

	t.Run("Use wildcards",
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "a.test.example.com"}}},
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
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "a.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "b.example.com"}}},
				{NameFQDN: "b.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "a.example.com"}}},
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}}},
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
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "a.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "a.example.com"}}},
				{NameFQDN: "www.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}}},
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
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "mail.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "mail.external.tld"}}},
			},
			[]string{
				"mail.example.com",
			},
			[]string{},
		),
	)

	t.Run("Deletions in correct order",
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "mail.example.com", Dependencies: []Dependency{}},
				{NameFQDN: "example.com", Dependencies: []Dependency{{NameFQDN: "mail.example.com", Type: BackwardDependency}}},
			},
			[]string{
				"example.com",
				"mail.example.com",
			},
			[]string{},
		),
	)

	t.Run("A Change with dependency on old and new state",
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "bar2.example.com", Dependencies: []Dependency{{}}},
				{NameFQDN: "foo.example.com", Dependencies: []Dependency{{NameFQDN: "bar2.example.com", Type: BackwardDependency}, {NameFQDN: "new2.example.com", Type: ForwardDependency}}},
				{NameFQDN: "new2.example.com", Dependencies: []Dependency{{}}},
			},
			[]string{
				"new2.example.com",
				"foo.example.com",
				"bar2.example.com",
			},
			[]string{},
		),
	)

	t.Run("Reports first and ignore dependencies",
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "example.com", Dependencies: []Dependency{}},
				{NameFQDN: "test.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}}, Type: Report},
			},
			[]string{
				"test.example.com",
				"example.com",
			},
			[]string{},
		),
	)

	t.Run("Reports are not resolvable as dependencies",
		executeGraphSort(
			[]stubRecord{
				{NameFQDN: "test2.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "test.example.com"}}},
				{NameFQDN: "test.example.com", Dependencies: []Dependency{}, Type: Report},
			},
			[]string{
				"test.example.com",
				"test2.example.com",
			},
			[]string{},
		),
	)
}

func executeGraphSort(inputOrder []stubRecord, expectedOutputOrder []string, expectedUnresolved []string) func(*testing.T) {
	return func(t *testing.T) {
		inputSortableRecords := stubRecordsAsSortableRecords(inputOrder)
		t.Helper()
		result := SortUsingGraph(inputSortableRecords)

		if len(expectedOutputOrder) != len(result.SortedRecords) {
			t.Errorf("Missing records after sort. Expected order: %v. Got order: %v\n", expectedOutputOrder, GetRecordsNamesForChanges(result.SortedRecords))
		} else {
			for iX := range expectedOutputOrder {
				if result.SortedRecords[iX].GetName() != expectedOutputOrder[iX] {
					t.Errorf("Invalid order on index %d after sort. Expected order: %v. Got order: %v\n", iX, expectedOutputOrder, GetRecordsNamesForChanges(result.SortedRecords))
				}
			}
		}

		if len(expectedUnresolved) != len(result.UnresolvedRecords) {
			t.Errorf("Missing unresolved records. Expected unresolved: %v. Got: %v\n", expectedUnresolved, GetRecordsNamesForChanges(result.UnresolvedRecords))
		} else {
			for iX := range expectedUnresolved {
				if result.UnresolvedRecords[iX].GetName() != expectedUnresolved[iX] {
					t.Errorf("Invalid unresolved records after sort. Expected: %v. Got: %v\n", expectedOutputOrder, GetRecordsNamesForChanges(result.UnresolvedRecords))
				}
			}
		}
	}
}
