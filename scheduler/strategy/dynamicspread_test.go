package strategy

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/scheduler/node"
	_ "github.com/docker/swarm/scheduler/notify" //add  notify package
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestName(t *testing.T) {
	var d = &DynamicSpreadPlacementStrategy{}
	var name = d.Name()
	assert.Equal(t, name, "dynamic-spread")
}

func err(has bool) error {
	if has {
		return errors.New("!!!")
	}

	return nil
}

func createConfig2(memory int64, cpus int64) *cluster.ContainerConfig {
	return cluster.BuildContainerConfig(dockerclient.ContainerConfig{Memory: memory * 1024 * 1024, CpuShares: cpus})
}

func TestRankAndSort2(t *testing.T) {
	var d = &DynamicSpreadPlacementStrategy{}

	var node1 = createNode("001", 4, 4)
	node1.JoinWeight = -1
	var nodes = []*node.Node{
		node1,
	}
	var config256 = createConfig2(256, 0)
	var _, err = d.RankAndSort(config256, nodes)
	assert.EqualError(t, NoHaveCustomNodes, err.Error())
}

func TestRankAndSort(t *testing.T) {
	var d = &DynamicSpreadPlacementStrategy{}

	var node1 = createNode("001", 4, 4)
	var node2 = createNode("002", 4, 4)
	var node3 = createNode("003", 4, 4)

	var nodes = []*node.Node{
		node1, node2, node3,
	}

	var config256 = createConfig2(256, 0)
	var config512 = createConfig2(512, 0)
	var config1024 = createConfig2(1024, 0)
	var config2048 = createConfig2(2048, 0)
	var config3072 = createConfig2(3072, 0)

	assert.NoError(t, node1.AddContainer(createContainer("c1", config256)))
	assert.NoError(t, node2.AddContainer(createContainer("c2", config512)))
	assert.Equal(t, node1.UsedMemory, int64(256*1024*1024))
	assert.Equal(t, node2.UsedMemory, int64(512*1024*1024))
	assert.Equal(t, node3.UsedMemory, int64(0))

	//node1 256  node2 512 node3 0
	var getNodes, err = d.RankAndSort(config256, nodes)
	assert.NoError(t, err)
	//测试平均调度
	assert.Equal(t, getNodes[0].ID, "003")
	assert.Equal(t, getNodes[1].ID, "001")
	assert.Equal(t, getNodes[2].ID, "002")

	assert.NoError(t, node3.AddContainer(createContainer("c3", config1024)))
	//node1 256  node2 512 node3 1024
	getNodes, err = d.RankAndSort(config256, nodes)
	assert.NoError(t, err)
	//测试平均调度
	assert.Equal(t, getNodes[0].ID, "001")
	assert.Equal(t, getNodes[1].ID, "002")
	assert.Equal(t, getNodes[2].ID, "003")

	assert.NoError(t, node1.AddContainer(createContainer("c4", config3072))) //3.25G
	assert.NoError(t, node2.AddContainer(createContainer("c5", config3072))) //3.5
	assert.NoError(t, node3.AddContainer(createContainer("c6", config2048))) //3G

	assert.NoError(t, node1.AddContainer(createContainer("c7", config512))) //3.75G
	assert.NoError(t, node2.AddContainer(createContainer("c8", config512))) //4G
	assert.NoError(t, node3.AddContainer(createContainer("c9", config512))) //3.5G
	// 3.5+4+3.75 /12 =0.93 应该发出通知了.
	log.Debugln("需要发邮件 ")
	getNodes, err = d.RankAndSort(config256, nodes)
	assert.NoError(t, err)
	assert.Equal(t, getNodes[0].ID, "003")
	assert.Equal(t, node1.UsedMemory, int64(3.75*1024*1024*1024))
	assert.Equal(t, node2.UsedMemory, int64(4*1024*1024*1024))
	assert.Equal(t, node3.UsedMemory, int64(3.5*1024*1024*1024))

	assert.NoError(t, node1.AddContainer(createContainer("c10", config256)))  //4G
	assert.NoError(t, node3.AddContainer(createContainer("c11", config1024))) //4.5G
	// node1 4G, node2  4G node3 4.5G 三台都已经满了
	assert.Equal(t, node1.UsedMemory, int64(4*1024*1024*1024))
	assert.Equal(t, node2.UsedMemory, int64(4*1024*1024*1024))
	assert.Equal(t, node3.UsedMemory, int64(4.5*1024*1024*1024))

	log.Debugln("需要发邮件 ,但是未到期.")
	getNodes, err = d.RankAndSort(config256, nodes) //发邮件,但未到期
	assert.NoError(t, err)
	assert.Equal(t, getNodes[0].ID, "002")
	assert.Equal(t, getNodes[1].ID, "001")

	log.Debugln("增加两台备用节点004,005")
	var bknode1 = createNode("004", 4, 4)
	var bknode2 = createNode("005", 4, 4)
	bknode1.JoinWeight = -1
	bknode2.JoinWeight = -1
	nodes = append(nodes, bknode1)
	nodes = append(nodes, bknode2) //增加两台备用节点

	getNodes, err = d.RankAndSort(config256, nodes) //有备用节点,不需要发邮件
	assert.NoError(t, err)
	assert.Contains(t, []string{"004", "005"}, getNodes[0].ID) //要么是004 或者是005
	assert.Contains(t, []string{"004", "005"}, getNodes[1].ID) //要么是004 或者是005

	assert.NoError(t, bknode1.AddContainer(createContainer("c12", config512))) //bk1 0.5G

	getNodes, err = d.RankAndSort(config256, nodes) //有备用节点,不需要发邮件
	assert.NoError(t, err)
	assert.Equal(t, getNodes[0].ID, "005")
	assert.Equal(t, getNodes[1].ID, "004")

	assert.NoError(t, bknode1.AddContainer(createContainer("c13", config3072))) //bk1 3.5G
	assert.NoError(t, bknode2.AddContainer(createContainer("c14", config3072))) //bk2 3G
	assert.NoError(t, bknode1.AddContainer(createContainer("c15", config512)))  //bk1 4G
	assert.NoError(t, bknode2.AddContainer(createContainer("c16", config512)))  //bk2 3.5G
	//4+3.5 / 8= 0.93
	getNodes, err = d.RankAndSort(config256, nodes) //集群超资源,提醒
	assert.NoError(t, err)
	assert.Equal(t, getNodes[0].ID, "005") //bk2

	// node1 4G, node2  4G node3 4.5G   bk1 4G    bk2 3.5G
	node1.RemoveContainer("c4")    //4-3=1G
	node3.RemoveContainer("c6")    //4.5-2=2.5G
	bknode2.RemoveContainer("c14") //3.5-3=0.5G
	// node1 1G, node2  4G node3 2G   bk1 4G    bk2 0.5G
	assert.Equal(t, node1.UsedMemory, int64(1*1024*1024*1024))
	assert.Equal(t, node2.UsedMemory, int64(4*1024*1024*1024))
	assert.Equal(t, node3.UsedMemory, int64(2.5*1024*1024*1024))
	assert.Equal(t, bknode1.UsedMemory, int64(4*1024*1024*1024))
	assert.Equal(t, bknode2.UsedMemory, int64(0.5*1024*1024*1024))

	getNodes, err = d.RankAndSort(config256, nodes) //常备节点处于可用状态
	assert.NoError(t, err)
	assert.Equal(t, getNodes[0].ID, "001") //node1
	assert.Equal(t, getNodes[1].ID, "003") //node3
	log.Info("延迟2s后退出...")
	time.Sleep(2 * time.Second)
}
