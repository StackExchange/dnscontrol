package domaintree

import (
	"strings"
)

type domaintree domainnode

type domainnode struct {
	IsLeaf     bool
	IsWildcard bool
	Name       string
	Children   map[string]*domainnode
}

func CreateTree() *domaintree {
	return &domaintree{
		IsLeaf:     false,
		IsWildcard: false,
		Name:       "",
		Children:   map[string]*domainnode{},
	}
}

func createNode(name string) *domainnode {
	return &domainnode{
		IsLeaf:   false,
		Name:     name,
		Children: map[string]*domainnode{},
	}
}

func (tree *domainnode) addIntermediate(name string) *domainnode {
	if _, ok := tree.Children[name]; !ok {
		tree.Children[name] = createNode(name)
	}

	return tree.Children[name]
}

func (tree *domainnode) addLeaf(name string, isWildcard bool) *domainnode {
	node := tree.addIntermediate(name)

	node.IsLeaf = true
	node.IsWildcard = node.IsWildcard || isWildcard

	return node
}

func (tree *domaintree) Add(domain string, name string) {
	fqdn := normalizeDomainName(domain, name)
	domainParts := strings.Split(fqdn, ".")

	isWildcard := domainParts[0] == "*"
	if isWildcard {
		domainParts = domainParts[1:]
	}

	ptr := (*domainnode)(tree)
	for iX := len(domainParts) - 1; iX > 0; iX -= 1 {
		ptr = ptr.addIntermediate(domainParts[iX])
	}

	ptr.addLeaf(domainParts[0], isWildcard)
}

func (tree *domaintree) Get(fqdn string) bool {
	domainParts := strings.Split(fqdn, ".")

	var mostSpecificNode *domainnode
	ptr := (*domainnode)(tree)

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

func normalizeDomainName(domain string, name string) string {
	domain = strings.TrimSuffix(domain, ".")
	if name == "@" {
		return domain
	}
	if strings.HasSuffix(name, ".") {
		return strings.TrimSuffix(name, ".")
	}

	return strings.TrimSuffix(name, ".") + "." + strings.TrimPrefix(domain, ".")
}
