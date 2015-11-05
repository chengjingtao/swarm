package cluster

import (
	_ "errors"
	_ "fmt"

	_ "strings"
	_ "sync"

	_ "github.com/Sirupsen/logrus"

	_ "github.com/samalba/dockerclient"
)

// UsedMemory returns the sum of memory reserved by containers.
func (e *Engine) UsedMemoryWithoutStop() int64 {
	var r int64
	e.RLock()
	for _, c := range e.containers {
		if c.Info.State.Running {
			r += c.Config.Memory
		}
	}
	e.RUnlock()
	return r
}

// UsedCpus returns the sum of CPUs reserved by containers.
func (e *Engine) UsedCpusWithoutStop() int64 {
	var r int64
	e.RLock()
	for _, c := range e.containers {
		if c.Info.State.Running {
			r += c.Config.CpuShares
		}
	}
	e.RUnlock()
	return r
}
