package target

// Target represents a target to be scanned
type Target struct {
	// Type of the target
	Type string
	// URL of the target
	URL string
}

// TargetContainer is a map of targets
type TargetContainer map[string]Target
