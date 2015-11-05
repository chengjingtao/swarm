package strategy

import (
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

// SpreadPlacementStrategy places a container on the node with the fewest running containers.
type ForeverSpreadPlacementStrategy struct {
}

// Initialize a SpreadPlacementStrategy.
func (p *ForeverSpreadPlacementStrategy) Initialize() error {
	return nil
}

// Name returns the name of the strategy.
func (p *ForeverSpreadPlacementStrategy) Name() string {
	return "forever-spread"
}

// RankAndSort sorts nodes based on the spread strategy applied to the container config.
func (p *ForeverSpreadPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error) {
	log.Debugln("forever-spread.RankAndSort")
	weightedNodes, err := weighAllNodes(config, nodes) //weighNodes(config, nodes)
	if err != nil {
		return nil, err
	}

	log.Debugln("sort nodes")
	sort.Sort(weightedNodes)
	output := make([]*node.Node, len(weightedNodes))
	for i, n := range weightedNodes {
		output[i] = n.Node
		log.Debugf("%d : %v , score : %d", i, n.Node.Addr, n.Weight)
	}
	return output, nil
}
