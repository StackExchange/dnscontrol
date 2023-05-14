package dnssort

import (
	"strings"
)

type DomainTree domainNode

type domainNode struct {
	IsLeaf     bool
	IsWildcard bool
	Name       string
	Children   map[string]*domainNode
}

func CreateTree() *DomainTree {
	return &DomainTree{
		IsLeaf:     false,
		IsWildcard: false,
		Name:       "",
		Children:   map[string]*domainNode{},
	}
}

func createNode(name string) *domainNode {
	return &domainNode{
		IsLeaf:   false,
		Name:     name,
		Children: map[string]*domainNode{},
	}
}

func (tree *domainNode) addIntermediate(name string) *domainNode {
	if _, ok := tree.Children[name]; !ok {
		tree.Children[name] = createNode(name)
	}

	return tree.Children[name]
}

func (tree *domainNode) addLeaf(name string, isWildcard bool) *domainNode {
	node := tree.addIntermediate(name)

	node.IsLeaf = true
	node.IsWildcard = node.IsWildcard || isWildcard

	return node
}

func (tree *DomainTree) Add(fqdn string) {
	domainParts := splitFQDN(fqdn)

	isWildcard := domainParts[0] == "*"
	if isWildcard {
		domainParts = domainParts[1:]
	}

	ptr := (*domainNode)(tree)
	for iX := len(domainParts) - 1; iX > 0; iX -= 1 {
		ptr = ptr.addIntermediate(domainParts[iX])
	}

	ptr.addLeaf(domainParts[0], isWildcard)
}

func (tree *DomainTree) Has(fqdn string) bool {
	domainParts := splitFQDN(fqdn)

	var mostSpecificNode *domainNode
	ptr := (*domainNode)(tree)

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
