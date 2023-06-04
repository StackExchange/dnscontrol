package dnssort

type ChangeType uint8

const (
	Change ChangeType = iota
	Report
)

type DependencyType uint8

const (
	ForwardDependency DependencyType = iota
	BackwardDependency
)

type Dependency struct {
	NameFQDN string
	Type     DependencyType
}

type SortableChange interface {
	GetType() ChangeType
	GetName() string
	GetDependencies() []Dependency
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
		names = append(names, change.GetName())
	}

	return names
}
