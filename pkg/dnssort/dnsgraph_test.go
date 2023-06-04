package dnssort

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateGraph(t *testing.T) {
	changes := []stubRecord{
		{NameFQDN: "example.com", Dependencies: []Dependency{}},
		{NameFQDN: "mail.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}, {Type: ForwardDependency, NameFQDN: "someserver.example.com"}}},
		{NameFQDN: "*.hq.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}}},
		{NameFQDN: "someserver.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "a.hq.example.com"}, {Type: ForwardDependency, NameFQDN: "b.hq.example.com"}}},
	}

	graph := CreateGraph(stubRecordsAsSortableRecords(changes))

	nodes := graph.tree.Get("example.com")
	assert.Len(t, nodes, 1)
	assert.Len(t, nodes[0].Edges, 2)
	assert.Equal(t, "mail.example.com", nodes[0].Edges[0].Node.Change.GetName())
	assert.Equal(t, IncomingEdge, nodes[0].Edges[0].Direction)
	assert.Equal(t, "*.hq.example.com", nodes[0].Edges[1].Node.Change.GetName())

	nodes = graph.tree.Get("someserver.example.com")
	assert.Len(t, nodes, 1)
	assert.Len(t, nodes[0].Edges, 3)
	assert.Equal(t, "mail.example.com", nodes[0].Edges[0].Node.Change.GetName())
	assert.Equal(t, IncomingEdge, nodes[0].Edges[0].Direction)
	assert.Equal(t, "*.hq.example.com", nodes[0].Edges[1].Node.Change.GetName())
	assert.Equal(t, OutgoingEdge, nodes[0].Edges[1].Direction)
	assert.Equal(t, "*.hq.example.com", nodes[0].Edges[2].Node.Change.GetName())
	assert.Equal(t, OutgoingEdge, nodes[0].Edges[2].Direction)

	nodes = graph.tree.Get("a.hq.example.com")
	assert.Len(t, nodes, 1)
	assert.Len(t, nodes[0].Edges, 3)
	assert.Equal(t, "example.com", nodes[0].Edges[0].Node.Change.GetName())
	assert.Equal(t, OutgoingEdge, nodes[0].Edges[0].Direction)
	assert.Equal(t, "someserver.example.com", nodes[0].Edges[1].Node.Change.GetName())
	assert.Equal(t, IncomingEdge, nodes[0].Edges[1].Direction)
	assert.Equal(t, "someserver.example.com", nodes[0].Edges[2].Node.Change.GetName())
	assert.Equal(t, IncomingEdge, nodes[0].Edges[2].Direction)
}

func Test_RemoveNode(t *testing.T) {
	changes := []stubRecord{
		{NameFQDN: "example.com", Dependencies: []Dependency{}},
		{NameFQDN: "mail.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}, {Type: ForwardDependency, NameFQDN: "someserver.example.com"}}},
		{NameFQDN: "*.hq.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "example.com"}}},
		{NameFQDN: "someserver.example.com", Dependencies: []Dependency{{Type: ForwardDependency, NameFQDN: "a.hq.example.com"}, {Type: ForwardDependency, NameFQDN: "b.hq.example.com"}}},
	}

	graph := CreateGraph(stubRecordsAsSortableRecords(changes))

	graph.removeNode(graph.tree.Get("example.com")[0])

	// example.com change has been removed
	nodes := graph.tree.Get("example.com")
	assert.Len(t, nodes, 0)

	nodes = graph.tree.Get("a.hq.example.com")
	assert.Len(t, nodes, 1)

	assert.Len(t, nodes[0].Edges, 2)
	assert.Equal(t, "someserver.example.com", nodes[0].Edges[0].Node.Change.GetName())
	assert.Equal(t, "someserver.example.com", nodes[0].Edges[1].Node.Change.GetName())
}
