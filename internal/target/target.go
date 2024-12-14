package target

import (
	"fmt"

	"github.com/ablankz/bloader/internal/config"
)

// Target represents a target to be scanned
type Target struct {
	// Type of the target
	Type config.TargetType
	// URL of the target
	URL string
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

// FindTarget returns a list of targets
func (tc TargetContainer) FindTarget(id string, targetType config.TargetType) (string, error) {
	if target, ok := tc[id]; ok {
		if target.Type == targetType {
			return target.URL, nil
		}
	}
	return "", fmt.Errorf("target(%s) not found on target(%s)", id, targetType)
}
