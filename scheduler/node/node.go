package node

import (
	"errors"

	"github.com/docker/swarm/cluster"
)

// Node is an abstract type used by the scheduler.
type Node struct {
	ID         string
	IP         string
	Addr       string
	Name       string
	Labels     map[string]string
	Containers cluster.Containers
	Images     []*cluster.Image

	UsedMemory  int64
	UsedCpus    int64
	TotalMemory int64
	TotalCpus   int64

	IsHealthy bool

	Containers_Start cluster.Containers
	JoinWeight       int //  weight  of join the cluster
}

// NewNode creates a node from an engine.
func NewNode(e *cluster.Engine) *Node {

	var _, start = e.ContainersTotalAndStart()

	return &Node{
		ID:               e.ID,
		IP:               e.IP,
		Addr:             e.Addr,
		Name:             e.Name,
		Labels:           e.Labels,
		Containers:       e.Containers(),
		Images:           e.Images(),
		UsedMemory:       e.UsedMemoryWithoutStop(), //e.UsedMemory(),
		UsedCpus:         e.UsedCpusWithoutStop(),   //e.UsedCpus(),
		TotalMemory:      e.TotalMemory(),
		TotalCpus:        e.TotalCpus(),
		IsHealthy:        e.IsHealthy(),
		Containers_Start: start,
		JoinWeight:       e.Weight,
	}
}

// Container returns the container with IDOrName in the engine.
func (n *Node) Container(IDOrName string) *cluster.Container {
	return n.Containers.Get(IDOrName)
}

// AddContainer injects a container into the internal state.
func (n *Node) AddContainer(container *cluster.Container) error {
	if container.Config != nil {
		memory := container.Config.Memory
		cpus := container.Config.CpuShares
		if n.TotalMemory-memory < 0 || n.TotalCpus-cpus < 0 {
			return errors.New("not enough resources")
		}
		n.UsedMemory = n.UsedMemory + memory
		n.UsedCpus = n.UsedCpus + cpus
	}
	n.Containers = append(n.Containers, container)
	return nil
}

// AddContainer injects a container into the internal state.
func (n *Node) RemoveContainer(id string) error {
	var newContainers = cluster.Containers{}

	for _, c := range n.Containers {
		if c.Id != id {
			newContainers = append(newContainers, c)
		} else {
			n.UsedMemory = n.UsedMemory - c.Config.Memory
			n.UsedCpus = n.UsedCpus - c.Config.CpuShares
		}
	}

	n.Containers = newContainers
	return nil
}
