package dnsgraph

func CreateDependencies(dependencyFQDNs []string, dependencyType DependencyType) []Dependency {
	var dependencies []Dependency

	for _, dependency := range dependencyFQDNs {
		dependencies = append(dependencies, Dependency{NameFQDN: dependency, Type: dependencyType})
	}

	return dependencies
}
