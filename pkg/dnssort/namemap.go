package dnssort

import "fmt"

type nameMap map[string]struct{}

func (names nameMap) add(name string) {
	names[name] = struct{}{}
}

func (names nameMap) has(name string) bool {
	_, ok := names[name]
	return ok
}

func (names nameMap) remove(name string) {
	delete(names, name)
}

func (names nameMap) empty() bool {
	return len(names) == 0
}

func (names nameMap) removeAll(resolvedNames *domaintree) {
	for dependency := range names {
		if resolvedNames.Has(dependency) {
			names.remove(dependency)
		}
	}
}

func (names nameMap) names() []string {
	nameList := []string{}
	for name := range names {
		nameList = append(nameList, name)
	}

	return nameList
}

func (names nameMap) String() string {
	return fmt.Sprintf("%v", names.names())
}
