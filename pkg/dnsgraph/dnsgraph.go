package dnsgraph

import "github.com/StackExchange/dnscontrol/v4/pkg/dnstree"

type edgeDirection uint8

const (
	IncomingEdge edgeDirection = iota
	OutgoingEdge
)

type DNSGraphEdge[T Graphable] struct {
	Dependency Dependency
	Node       *DNSGraphNode[T]
	Direction  edgeDirection
}

type DNSGraphEdges[T Graphable] []DNSGraphEdge[T]

type DNSGraphNode[T Graphable] struct {
	Data  T
	Edges DNSGraphEdges[T]
}

type dnsGraphNodes[T Graphable] []*DNSGraphNode[T]

type DNSGraph[T Graphable] struct {
	All  dnsGraphNodes[T]
	Tree *dnstree.DomainTree[dnsGraphNodes[T]]
}

func CreateGraph[T Graphable](entries []T) *DNSGraph[T] {
	graph := &DNSGraph[T]{
		All:  dnsGraphNodes[T]{},
		Tree: dnstree.Create[dnsGraphNodes[T]](),
	}

	for _, data := range entries {
		graph.AddNode(data)
	}

	for _, data := range entries {
		sourceNodes := graph.Tree.Get(data.GetName())
		for _, dependency := range data.GetDependencies() {
			graph.AddEdge(sourceNodes, dependency)
		}
	}

	return graph
}

func (graph *DNSGraph[T]) RemoveNode(toRemove *DNSGraphNode[T]) {
	for _, edge := range toRemove.Edges {
		edge.Node.Edges = edge.Node.Edges.RemoveNode(toRemove)
	}

	graph.All = graph.All.RemoveNode(toRemove)

	nodes := graph.Tree.Get(toRemove.Data.GetName())
	if nodes != nil {
		nodes = nodes.RemoveNode(toRemove)
		graph.Tree.Set(toRemove.Data.GetName(), nodes)
	}
}

func (graph *DNSGraph[T]) AddNode(data T) {
	nodes := graph.Tree.Get(data.GetName())
	node := &DNSGraphNode[T]{
		Data:  data,
		Edges: DNSGraphEdges[T]{},
	}
	if nodes == nil {
		nodes = dnsGraphNodes[T]{node}
	}

	graph.All = append(graph.All, node)
	graph.Tree.Set(data.GetName(), nodes)
}

func (graph *DNSGraph[T]) AddEdge(sourceNodes []*DNSGraphNode[T], dependency Dependency) {
	destinationNodes := graph.Tree.Get(dependency.NameFQDN)

	if sourceNodes == nil || destinationNodes == nil {
		return
	}

	for _, sourceNode := range sourceNodes {
		for _, destinationNode := range destinationNodes {
			sourceNode.Edges = append(sourceNode.Edges, DNSGraphEdge[T]{
				Dependency: dependency,
				Node:       destinationNode,
				Direction:  OutgoingEdge,
			})

			destinationNode.Edges = append(destinationNode.Edges, DNSGraphEdge[T]{
				Dependency: dependency,
				Node:       sourceNode,
				Direction:  IncomingEdge,
			})
		}
	}
}

func (nodes dnsGraphNodes[T]) RemoveNode(toRemove *DNSGraphNode[T]) dnsGraphNodes[T] {
	var newNodes dnsGraphNodes[T]

	for _, node := range nodes {
		if node != toRemove {
			newNodes = append(newNodes, node)
		}
	}

	return newNodes
}

func (edges DNSGraphEdges[T]) RemoveNode(toRemove *DNSGraphNode[T]) DNSGraphEdges[T] {
	var newEdges DNSGraphEdges[T]

	for _, edge := range edges {
		if edge.Node != toRemove {
			newEdges = append(newEdges, edge)
		}
	}

	return newEdges
}
