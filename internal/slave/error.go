package slave

import "fmt"

var (
	// ErrInvalidEnvironment represents an error when the environment is invalid
	ErrInvalidEnvironment = fmt.Errorf("must connect to the same environment")
)
