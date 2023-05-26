package dnssort

type ChangeType uint8

const (
	CHANGE ChangeType = iota
	REPORT
)

type DependencyType uint8

const (
	NEW_DEPENDENCY DependencyType = iota
	OLD_DEPENDENCY
)

type Dependency struct {
	NameFQDN string
	Type     DependencyType
}

type SortableChange interface {
	GetType() ChangeType
	GetNameFQDN() string
	GetFQDNDependencies() []Dependency
	Equals(change SortableChange) bool
}

func CreateDependencies(dependencyFQDNs []string, dependencyType DependencyType) []Dependency {
	var dependencies []Dependency

	for _, dependency := range dependencyFQDNs {
		dependencies = append(dependencies, Dependency{NameFQDN: dependency, Type: dependencyType})
	}

	return dependencies
}
