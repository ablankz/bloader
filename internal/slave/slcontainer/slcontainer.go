package slcontainer

import (
	"sync"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/target"
)

// SlaveContainer represents the container for the slave node
type SlaveContainer struct {
	mu     *sync.RWMutex
	Target Target
}

// NewSlaveContainer creates a new container for the slave node
func NewSlaveContainer(ctr *container.Container) *SlaveContainer {
	return &SlaveContainer{
		Target: Target{
			TargetContainer: make(map[string]target.Target),
		},
	}
}
