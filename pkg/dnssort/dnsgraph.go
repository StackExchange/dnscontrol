package dnssort

type dnsGraphNode struct {
	change   SortableChange
	incoming dnsGraphNodes
	outgoing dnsGraphNodes
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
		for _, dependency := range change.GetFQDNDependencies() {
			graph.addEdge(change.GetNameFQDN(), dependency)
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

	for _, incoming := range toRemove.incoming {
		incoming.outgoing = incoming.outgoing.removeNode(toRemove)
	}

	for _, outgoing := range toRemove.outgoing {
		outgoing.incoming = outgoing.incoming.removeNode(toRemove)
	}

	if index > -1 {
		graph.all = append(graph.all[:index], graph.all[index+1:]...)
		nodes := graph.tree.Get(toRemove.change.GetNameFQDN())
		nodes = nodes.removeNode(toRemove)
		graph.tree.Add(toRemove.change.GetNameFQDN(), nodes)
	}
}

func (graph *DNSGraph) addNode(change SortableChange) {
	nodes := graph.tree.Get(change.GetNameFQDN())
	node := &dnsGraphNode{
		change:   change,
		incoming: dnsGraphNodes{},
		outgoing: dnsGraphNodes{},
	}
	if nodes == nil {
		nodes = dnsGraphNodes{node}
	}

	graph.all = append(graph.all, node)
	graph.tree.Add(change.GetNameFQDN(), nodes)
}

func (graph *DNSGraph) addEdge(source string, destination string) {
	sourceNodes := graph.tree.Get(source)
	destinationNodes := graph.tree.Get(destination)

	if sourceNodes == nil || destinationNodes == nil {
		return
	}

	for _, sourceNode := range sourceNodes {
		for _, destinationNode := range destinationNodes {
			sourceNode.outgoing = append(sourceNode.outgoing, destinationNode)
			destinationNode.incoming = append(destinationNode.incoming, sourceNode)
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
