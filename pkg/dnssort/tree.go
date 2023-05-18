package dnssort

import (
	"strings"
)

type DomainTree[T any] domainNode[T]

type domainNode[T any] struct {
	IsLeaf     bool
	IsWildcard bool
	Name       string
	Children   map[string]*domainNode[T]
	data       T
}

func CreateTree[T any]() *DomainTree[T] {
	return &DomainTree[T]{
		IsLeaf:     false,
		IsWildcard: false,
		Name:       "",
		Children:   map[string]*domainNode[T]{},
	}
}

func createNode[T any](name string) *domainNode[T] {
	return &domainNode[T]{
		IsLeaf:   false,
		Name:     name,
		Children: map[string]*domainNode[T]{},
	}
}

func (tree *domainNode[T]) addIntermediate(name string) *domainNode[T] {
	if _, ok := tree.Children[name]; !ok {
		tree.Children[name] = createNode[T](name)
	}

	return tree.Children[name]
}

func (tree *domainNode[T]) addLeaf(name string, isWildcard bool, data T) *domainNode[T] {
	node := tree.addIntermediate(name)

	node.data = data
	node.IsLeaf = true
	node.IsWildcard = node.IsWildcard || isWildcard

	return node
}

func (tree *DomainTree[T]) Add(fqdn string, data T) {
	domainParts := splitFQDN(fqdn)

	isWildcard := domainParts[0] == "*"
	if isWildcard {
		domainParts = domainParts[1:]
	}

	ptr := (*domainNode[T])(tree)
	for iX := len(domainParts) - 1; iX > 0; iX -= 1 {
		ptr = ptr.addIntermediate(domainParts[iX])
	}

	ptr.addLeaf(domainParts[0], isWildcard, data)
}

func (tree *DomainTree[T]) Get(fqdn string) T {
	domainParts := splitFQDN(fqdn)

	var mostSpecificNode *domainNode[T]
	ptr := (*domainNode[T])(tree)

	for iX := len(domainParts) - 1; iX >= 0; iX -= 1 {
		node, ok := ptr.Children[domainParts[iX]]
		if !ok {
			if mostSpecificNode != nil {
				return mostSpecificNode.data
			}
			return *new(T)
		}

		if node.IsWildcard {
			mostSpecificNode = node
		}

		ptr = node
	}

	if ptr.IsLeaf || ptr.IsWildcard {
		return ptr.data
	}

	if mostSpecificNode != nil {
		return mostSpecificNode.data
	}

	return *new(T)
}

func (tree *DomainTree[T]) Has(fqdn string) bool {
	domainParts := splitFQDN(fqdn)

	var mostSpecificNode *domainNode[T]
	ptr := (*domainNode[T])(tree)

	for iX := len(domainParts) - 1; iX >= 0; iX -= 1 {
		node, ok := ptr.Children[domainParts[iX]]
		if !ok {
			return mostSpecificNode != nil
		}

		if node.IsWildcard {
			mostSpecificNode = node
		}

		ptr = node
	}

	return ptr.IsLeaf || ptr.IsWildcard || mostSpecificNode != nil
}

func splitFQDN(fqdn string) []string {
	normalizedFQDN := strings.TrimSuffix(fqdn, ".")

	return strings.Split(normalizedFQDN, ".")
}
