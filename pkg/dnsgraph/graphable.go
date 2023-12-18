package dnsgraph

// NodeType enumerates the node types.
type NodeType uint8

const (
	// Change is the type of change.
	Change NodeType = iota
	// Report is a Report.
	Report
)

// DependencyType enumerates the dependency types.
type DependencyType uint8

const (
	// ForwardDependency is a forward dependency.
	ForwardDependency DependencyType = iota
	// BackwardDependency is a backwards dependency.
	BackwardDependency
)

// Dependency is a dependency.
type Dependency struct {
	NameFQDN string
	Type     DependencyType
}

// Graphable is an interface for things that can be in a graph.
type Graphable interface {
	GetType() NodeType
	GetName() string
	GetDependencies() []Dependency
}

// GetRecordsNamesForGraphables returns names in a graph.
func GetRecordsNamesForGraphables[T Graphable](graphables []T) []string {
	var names []string

	for _, graphable := range graphables {
		names = append(names, graphable.GetName())
	}

	return names
}
