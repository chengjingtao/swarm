package strategy

import (
	"errors"
	_ "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
)

func weighAllNodes(config *cluster.ContainerConfig, nodes []*node.Node) (weightedNodeList, error) {
	//log.Debugln("=============weightAllNodes===========")

	weightedNodes := weightedNodeList{}

	var weightItem = func(item *node.Node) *weightedNode {
		var (
			cpuScore    int64 = 100
			memoryScore int64 = 100
		)

		if config.CpuShares >= 0 {
			cpuScore = (item.UsedCpus + config.CpuShares) * 100 / item.TotalCpus
		}
		if config.Memory >= 0 {
			memoryScore = (item.UsedMemory + config.Memory) * 100 / item.TotalMemory
		}

		//log.Debugf("node.Addr:%s ,  node.TotalCpus:%d , node.TotalMemory:%d ,   config.Cpu:%d   ,   config.Memory:%d  ,  usedMemory :%d ,  usedCpu: %d ,   CpuScore:%d  ,  MemoryScore:%d",
		//	item.Addr, item.TotalCpus, item.TotalMemory, config.CpuShares, config.Memory,
		//	item.UsedMemory+config.Memory, item.UsedCpus+config.CpuShares, cpuScore, memoryScore)

		return &weightedNode{Node: item, Weight: cpuScore + memoryScore, memoryScore: memoryScore, cpuScore: cpuScore}
	}

	for _, node := range nodes {

		weighted := weightItem(node)

		if node.TotalMemory < int64(config.Memory) || node.TotalCpus < config.CpuShares {
			weighted.Weight += 200 //直接加上200 保证大于那些有可用资源的服务器排序后排在前边.
		}

		weightedNodes = append(weightedNodes, weighted)
	}

	if len(weightedNodes) == 0 {
		return nil, errors.New("no nodes in cluster")
	}

	return weightedNodes, nil
}
