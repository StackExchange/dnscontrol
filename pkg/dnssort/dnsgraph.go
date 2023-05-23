package dnssort

type edgeDirection uint8

const (
	IncomingEdge edgeDirection = iota
	OutgoingEdge
)

type dnsGraphEdge struct {
	Dependency Dependency
	Node       *dnsGraphNode
	Direction  edgeDirection
}

type dnsGraphEdges []dnsGraphEdge

type dnsGraphNode struct {
	Change SortableChange
	Edges  dnsGraphEdges
}

type dnsGraphNodes []*dnsGraphNode

type DNSGraph struct {
	all  dnsGraphNodes
	tree *DomainTree[dnsGraphNodes]
}

func CreateGraph(changes []SortableChange) *DNSGraph {
	graph := &DNSGraph{
		all:  dnsGraphNodes{},
		tree: CreateTree[dnsGraphNodes](),
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

func (graph *DNSGraph) removeNode(toRemove *dnsGraphNode) {
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
		graph.all = append(graph.all[:index], graph.all[index+1:]...)
		nodes := graph.tree.Get(toRemove.Change.GetNameFQDN())
		nodes = nodes.removeNode(toRemove)
		graph.tree.Set(toRemove.Change.GetNameFQDN(), nodes)
	}
}

func (graph *DNSGraph) addNode(change SortableChange) {
	nodes := graph.tree.Get(change.GetNameFQDN())
	node := &dnsGraphNode{
		Change: change,
		Edges:  dnsGraphEdges{},
	}
	if nodes == nil {
		nodes = dnsGraphNodes{node}
	}

	graph.all = append(graph.all, node)
	graph.tree.Set(change.GetNameFQDN(), nodes)
}

func (graph *DNSGraph) addEdge(sourceNodes []*dnsGraphNode, dependency Dependency) {
	destinationNodes := graph.tree.Get(dependency.NameFQDN)

	if sourceNodes == nil || destinationNodes == nil {
		return
	}

	for _, sourceNode := range sourceNodes {
		for _, destinationNode := range destinationNodes {
			sourceNode.Edges = append(sourceNode.Edges, dnsGraphEdge{
				Dependency: dependency,
				Node:       destinationNode,
				Direction:  OutgoingEdge,
			})

			destinationNode.Edges = append(destinationNode.Edges, dnsGraphEdge{
				Dependency: dependency,
				Node:       sourceNode,
				Direction:  IncomingEdge,
			})
		}
	}
}

func (nodes dnsGraphNodes) removeNode(toRemove *dnsGraphNode) dnsGraphNodes {
	var newNodes dnsGraphNodes

	for _, node := range nodes {
		if node != toRemove {
			newNodes = append(newNodes, node)
		}
	}

	return newNodes
}

func (edges dnsGraphEdges) removeNode(toRemove *dnsGraphNode) dnsGraphEdges {
	var newEdges dnsGraphEdges

	for _, edge := range edges {
		if edge.Node != toRemove {
			newEdges = append(newEdges, edge)
		}
	}

	return newEdges
}
