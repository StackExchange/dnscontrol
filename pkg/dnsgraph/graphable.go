package dnsgraph

type NodeType uint8

const (
	Change NodeType = iota
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

type Graphable interface {
	GetType() NodeType
	GetName() string
	GetDependencies() []Dependency
}

func GetRecordsNamesForGraphables[T Graphable](graphables []T) []string {
	var names []string

	for _, graphable := range graphables {
		names = append(names, graphable.GetName())
	}

	return names
}
