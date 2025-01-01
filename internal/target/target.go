package target

import (
	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"

	"github.com/ablankz/bloader/internal/config"
)

// Target represents a target to be scanned
type Target struct {
	// Type of the target
	Type config.TargetType
	// URL of the target
	URL string
}

// GetTarget returns the target
func (t Target) GetTarget() *pb.Target {
	switch t.Type {
	case config.TargetTypeHTTP:
		return &pb.Target{
			Type: pb.TargetType_TARGET_TYPE_HTTP,
			Target: &pb.Target_Http{
				Http: &pb.TargetHTTPData{
					Url: t.URL,
				},
			},
		}
	}

	return nil
}

// TargetContainer is a map of targets
type TargetContainer map[string]Target

// NewTargetContainer creates a new TargetContainer
func NewTargetContainer(env string, cfg config.ValidTargetConfig) TargetContainer {
	targets := make(TargetContainer)
	for _, target := range cfg {
		t := Target{
			Type: target.Type,
		}
		var ok bool
		for _, val := range target.Values {
			if val.Env == env {
				t.URL = val.URL
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		targets[target.ID] = t
	}
	return targets
}
