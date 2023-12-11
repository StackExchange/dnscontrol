package dnsgraph

// CreateDependencies creates a dependency list from a list of FQDNs and a DepenendyType.
func CreateDependencies(dependencyFQDNs []string, dependencyType DependencyType) []Dependency {
	var dependencies []Dependency

	for _, dependency := range dependencyFQDNs {
		dependencies = append(dependencies, Dependency{NameFQDN: dependency, Type: dependencyType})
	}

	return dependencies
}
