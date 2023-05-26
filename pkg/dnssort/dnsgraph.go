package dnssort

type edgeDirection uint8

const (
	IncomingEdge edgeDirection = iota
	OutgoingEdge
)

type dnsGraphEdge[T SortableChange] struct {
	Dependency Dependency
	Node       *dnsGraphNode[T]
	Direction  edgeDirection
}

type dnsGraphEdges[T SortableChange] []dnsGraphEdge[T]

type dnsGraphNode[T SortableChange] struct {
	Change T
	Edges  dnsGraphEdges[T]
}

type dnsGraphNodes[T SortableChange] []*dnsGraphNode[T]

type DNSGraph[T SortableChange] struct {
	all  dnsGraphNodes[T]
	tree *DomainTree[dnsGraphNodes[T]]
}

func CreateGraph[T SortableChange](changes []T) *DNSGraph[T] {
	graph := &DNSGraph[T]{
		all:  dnsGraphNodes[T]{},
		tree: CreateTree[dnsGraphNodes[T]](),
	}

	for _, change := range changes {
		graph.addNode(change)
	}

	for _, change := range changes {
		sourceNodes := graph.tree.Get(change.GetNameFQDN())
		for _, dependency := range change.GetFQDNDependencies() {
			graph.addEdge(sourceNodes, dependency)
		}
	}

	return graph
}

func (graph *DNSGraph[T]) removeNode(toRemove *dnsGraphNode[T]) {
	index := -1
	for iX, node := range graph.all {
		if node == toRemove {
			index = iX
			break
		}
	}

	for _, edge := range toRemove.Edges {
		edge.Node.Edges = edge.Node.Edges.removeNode(toRemove)
	}

	if index > -1 {
		graph.all[index] = graph.all[len(graph.all)-1]
		graph.all = graph.all[:len(graph.all)-1]
		nodes := graph.tree.Get(toRemove.Change.GetNameFQDN())
		nodes = nodes.removeNode(toRemove)
		graph.tree.Set(toRemove.Change.GetNameFQDN(), nodes)
	}
}

func (graph *DNSGraph[T]) addNode(change T) {
	nodes := graph.tree.Get(change.GetNameFQDN())
	node := &dnsGraphNode[T]{
		Change: change,
		Edges:  dnsGraphEdges[T]{},
	}
	if nodes == nil {
		nodes = dnsGraphNodes[T]{node}
	}

	graph.all = append(graph.all, node)
	graph.tree.Set(change.GetNameFQDN(), nodes)
}

func (graph *DNSGraph[T]) addEdge(sourceNodes []*dnsGraphNode[T], dependency Dependency) {
	destinationNodes := graph.tree.Get(dependency.NameFQDN)

	if sourceNodes == nil || destinationNodes == nil {
		return
	}

	for _, sourceNode := range sourceNodes {
		for _, destinationNode := range destinationNodes {
			sourceNode.Edges = append(sourceNode.Edges, dnsGraphEdge[T]{
				Dependency: dependency,
				Node:       destinationNode,
				Direction:  OutgoingEdge,
			})

			destinationNode.Edges = append(destinationNode.Edges, dnsGraphEdge[T]{
				Dependency: dependency,
				Node:       sourceNode,
				Direction:  IncomingEdge,
			})
		}
	}
}

func (nodes dnsGraphNodes[T]) removeNode(toRemove *dnsGraphNode[T]) dnsGraphNodes[T] {
	var newNodes dnsGraphNodes[T]

	for _, node := range nodes {
		if node != toRemove {
			newNodes = append(newNodes, node)
		}
	}

	return newNodes
}

func (edges dnsGraphEdges[T]) removeNode(toRemove *dnsGraphNode[T]) dnsGraphEdges[T] {
	var newEdges dnsGraphEdges[T]

	for _, edge := range edges {
		if edge.Node != toRemove {
			newEdges = append(newEdges, edge)
		}
	}

	return newEdges
}
