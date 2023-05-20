package dnssort

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateGraph(t *testing.T) {
	changes := []stubRecord{
		{NameFQDN: "example.com", Dependencies: []Dependency{}},
		{NameFQDN: "mail.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}, {Type: NewDependency, NameFQDN: "someserver.example.com"}}},
		{NameFQDN: "*.hq.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}}},
		{NameFQDN: "someserver.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "a.hq.example.com"}, {Type: NewDependency, NameFQDN: "b.hq.example.com"}}},
	}

	graph := CreateGraph(stubRecordsAsSortableRecords(changes))

	nodes := graph.tree.Get("example.com")
	assert.Len(t, nodes, 1)
	assert.Len(t, nodes[0].outgoing, 0)
	assert.Len(t, nodes[0].incoming, 2)
	assert.Equal(t, "mail.example.com", nodes[0].incoming[0].change.GetNameFQDN())
	assert.Equal(t, "*.hq.example.com", nodes[0].incoming[1].change.GetNameFQDN())

	nodes = graph.tree.Get("someserver.example.com")
	assert.Len(t, nodes, 1)
	assert.Len(t, nodes[0].outgoing, 2)
	assert.Len(t, nodes[0].incoming, 1)
	assert.Equal(t, "*.hq.example.com", nodes[0].outgoing[0].change.GetNameFQDN())
	assert.Equal(t, "*.hq.example.com", nodes[0].outgoing[1].change.GetNameFQDN())
	assert.Equal(t, "mail.example.com", nodes[0].incoming[0].change.GetNameFQDN())

	nodes = graph.tree.Get("a.hq.example.com")
	assert.Len(t, nodes, 1)
	assert.Len(t, nodes[0].outgoing, 1)
	assert.Len(t, nodes[0].incoming, 2)
	assert.Equal(t, "someserver.example.com", nodes[0].incoming[0].change.GetNameFQDN())
	assert.Equal(t, "someserver.example.com", nodes[0].incoming[1].change.GetNameFQDN())
	assert.Equal(t, "example.com", nodes[0].outgoing[0].change.GetNameFQDN())
}

func Test_RemoveNode(t *testing.T) {
	changes := []stubRecord{
		{NameFQDN: "example.com", Dependencies: []Dependency{}},
		{NameFQDN: "mail.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}, {Type: NewDependency, NameFQDN: "someserver.example.com"}}},
		{NameFQDN: "*.hq.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "example.com"}}},
		{NameFQDN: "someserver.example.com", Dependencies: []Dependency{{Type: NewDependency, NameFQDN: "a.hq.example.com"}, {Type: NewDependency, NameFQDN: "b.hq.example.com"}}},
	}

	graph := CreateGraph(stubRecordsAsSortableRecords(changes))

	graph.removeNode(graph.tree.Get("example.com")[0])

	// example.com change has been removed
	nodes := graph.tree.Get("example.com")
	assert.Len(t, nodes, 0)

	nodes = graph.tree.Get("a.hq.example.com")
	assert.Len(t, nodes, 1)

	// Outgoing has been removed
	assert.Len(t, nodes[0].outgoing, 0)

	// Incoming remains untouched
	assert.Len(t, nodes[0].incoming, 2)
	assert.Equal(t, "someserver.example.com", nodes[0].incoming[0].change.GetNameFQDN())
	assert.Equal(t, "someserver.example.com", nodes[0].incoming[1].change.GetNameFQDN())
}
