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
	Equals(change SortableChange) bool
}
