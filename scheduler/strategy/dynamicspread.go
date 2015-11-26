package strategy

import (
	"errors"
	"fmt"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/eventemitter"
	"github.com/docker/swarm/scheduler/node"
)

// SpreadPlacementStrategy places a container on the node with the fewest running containers.
type DynamicSpreadPlacementStrategy struct {
}

// Initialize a SpreadPlacementStrategy.
func (p *DynamicSpreadPlacementStrategy) Initialize() error {
	return nil
}

// Name returns the name of the strategy.
func (p *DynamicSpreadPlacementStrategy) Name() string {
	return "dynamic-spread"
}

// RankAndSort sorts nodes based on the spread strategy applied to the container config.
func (p *DynamicSpreadPlacementStrategy) RankAndSort(config *cluster.ContainerConfig, nodes []*node.Node) ([]*node.Node, error) {
	weightedList, err := weighAllNodes(config, nodes)

	if err != nil {
		return nil, err
	}
	sort.Sort(weightedList)

	var notBackupNodeList weightedNodeList
	var backupNodeList weightedNodeList
	var noOverInNotBkList weightedNodeList

	var memoryLoad float32
	var cpuLoad float32

	var memoryLoadBk float32
	var cpuLoadBk float32

	for _, item := range weightedList {
		if item.Node.JoinWeight == -1 {
			backupNodeList = append(backupNodeList, item)

			memoryLoadBk += item.MemoryLoad()
			cpuLoadBk += item.CpuLoad()

		} else { //not back up list
			notBackupNodeList = append(notBackupNodeList, item)

			memoryLoad += item.MemoryLoad()
			cpuLoad += item.CpuLoad()
			//log.Debugf("%s memload=%.2f cpuload=%.2f", item.Node.ID, memoryLoad, cpuLoad)
			//资源没有超限
			if item.MemoryLoad() < 1 && item.CpuLoad() < 1 {
				noOverInNotBkList = append(noOverInNotBkList, item)
			}
		}
	}

	if len(notBackupNodeList) <= 0 {
		return nil, errors.New("常备节点为空,无法调度.")
	}

	var avgMemLoad = memoryLoad / float32(len(notBackupNodeList))
	var avgCpuLoad = cpuLoad / float32(len(notBackupNodeList))

	//没有备用服务器的情况下
	if len(backupNodeList) == 0 {
		log.Debugf("无备用服务器 ,avg-memory = %.2f  avg-cpu = %.2f", avgMemLoad, avgCpuLoad)
		//内存或者cpu负载超过90%
		if avgMemLoad > 0.9 || avgCpuLoad > 0.9 {
			var msg = fmt.Sprintf("集群常备节点内存负载为 %.2f , cpu 负载为 %.2f", avgMemLoad, avgCpuLoad)
			eventemitter.Emit("/cluster/resources/over", []interface{}{msg, getNodes(weightedList)})
		}
		return getNodes(notBackupNodeList), nil
	} else { //存在备用服务器

		var avgMemLoadBk = memoryLoadBk / float32(len(backupNodeList))
		var avgCpuLoadBk = cpuLoadBk / float32(len(backupNodeList))
		log.Debugf("存在备用服务器 ,avg-memory-bk = %.2f  avg-cpu-bk = %.2f", avgMemLoadBk, avgCpuLoadBk)
		//备用服务器和常备服务器的资源都超过90%,发出提醒
		if (avgMemLoad > 0.9 || avgCpuLoad > 0.9) && (avgMemLoadBk > 0.9 || avgCpuLoadBk > 0.9) {
			var msg = fmt.Sprintf("集群所有节点内存负载为 %.2f , cpu 负载为 %.2f", avgMemLoad, avgCpuLoad)
			log.Debugln(msg)
			eventemitter.Emit("/cluster/resources/over", []interface{}{msg, getNodes(weightedList)})
		}

		//常备服务器中无资源尚有资源可用 优先使用常备服务器
		if len(noOverInNotBkList) > 0 {
			return getNodes(noOverInNotBkList), nil
		}

		//常备服务器中无资源可用,则在所有服务器中挑出 排名靠前者
		return getNodes(weightedList), nil
	}
}

func getNodes(wNodes weightedNodeList) []*node.Node {
	var nodes = []*node.Node{}
	for _, item := range wNodes {
		nodes = append(nodes, item.Node)
	}

	return nodes
}
