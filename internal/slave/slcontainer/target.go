package slcontainer

import (
	"fmt"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/target"
)

// Target represents the target container for the slave node
type Target struct {
	target.TargetContainer
}

// Exists checks if the target exists
func (t Target) Exists(id string) bool {
	_, ok := t.TargetContainer[id]
	return ok
}

// Add adds a new target to the container
func (t *Target) Add(id string, target target.Target) {
	t.TargetContainer[id] = target
}

// Remove removes a target from the container
func (t *Target) Remove(id string) {
	delete(t.TargetContainer, id)
}

// AddFromProto adds a new target from the proto to the container
func (t Target) AddFromProto(id string, pbT *pb.Target) error {
	switch pbT.Type {
	case pb.TargetType_TARGET_TYPE_HTTP:
		t.Add(id, target.Target{
			Type: config.TargetTypeHTTP,
			URL:  pbT.GetHttp().Url,
		})
	case pb.TargetType_TARGET_TYPE_UNSPECIFIED:
		return fmt.Errorf("invalid target type: %v", pbT.Type)
	}

	return fmt.Errorf("invalid target type: %v", pbT.Type)
}
