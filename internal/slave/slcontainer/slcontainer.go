package slcontainer

import "github.com/ablankz/bloader/internal/container"

// SlaveContainer represents the container for the slave node
type SlaveContainer struct {
}

// NewSlaveContainer creates a new container for the slave node
func NewSlaveContainer(ctr *container.Container) *SlaveContainer {
	return &SlaveContainer{}
}
