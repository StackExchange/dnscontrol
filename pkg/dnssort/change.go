package dnssort

type ChangeType uint8

const (
	AdditionChange ChangeType = iota
	DeletionChange
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
