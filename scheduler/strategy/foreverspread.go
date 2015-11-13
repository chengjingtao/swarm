package strategy

import (
	"errors"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/eventemitter"
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

	if len(weightedNodes) == 0 {
		return nil, errors.New("Impossible Happen. no nodes in cluster")
	}

	//暂时放在这个地方吧...
	log.Debugln("NOTIFY JUDGE")
	if notify := isNeedToNotify(config, nodes); notify != "" {
		//evt:/cluster/resources/over, args: []*node.Node
		eventemitter.Emit("/cluster/resources/over", []interface{}{notify, output})
	}

	return output, nil
}

//是否需要提醒目前来看和调度没有耦合关系, 如果 需要提醒虽然意味这没有可用资源,但是并不代表 PlaceContainer失败,所以目前将两者分开.
//如果要做在一起,那么就需要重构RankAndSort的含义,改造所有策略
func isNeedToNotify(config *cluster.ContainerConfig, nodes []*node.Node) string {
	var cpuOver = true
	var memOver = true

	for _, node := range nodes {
		var willCpus = node.UsedCpus + config.CpuShares
		var willMemory = node.UsedMemory + config.Memory

		if willCpus < node.TotalCpus {
			cpuOver = false //有一个没有超过
		}

		if willMemory < node.TotalMemory {
			memOver = false
		}

		log.Debugf("NODE[%s] OVER= %v  CPU = %d / %d   Memory = %d / %d", node.Addr,
			cpuOver || memOver,
			willCpus, node.TotalCpus, willMemory, node.TotalMemory)

	}

	if cpuOver && memOver {
		return "cpu && memory"
	}

	if cpuOver {
		return "cpu"
	}

	if memOver {
		return "memory"
	}

	return ""
}
