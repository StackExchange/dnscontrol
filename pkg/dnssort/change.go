package dnssort

type ChangeType uint8

const (
	Change ChangeType = iota
	Report
)

type DependencyType uint8

const (
	NewDependency DependencyType = iota
	OldDependency
)

type Dependency struct {
	NameFQDN string
	Type     DependencyType
}

type SortableChange interface {
	GetType() ChangeType
	GetNameFQDN() string
	GetFQDNDependencies() []Dependency
}

func CreateDependencies(dependencyFQDNs []string, dependencyType DependencyType) []Dependency {
	var dependencies []Dependency

	for _, dependency := range dependencyFQDNs {
		dependencies = append(dependencies, Dependency{NameFQDN: dependency, Type: dependencyType})
	}

	return dependencies
}

func GetRecordsNamesForChanges[T SortableChange](changes []T) []string {
	var names []string

	for _, change := range changes {
		names = append(names, change.GetNameFQDN())
	}

	return names
}
