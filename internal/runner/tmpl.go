package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
)

// TmplFactor represents the template factor
type TmplFactor interface {
	// TmplFactorize returns the factorized template
	TmplFactorize(ctx context.Context, path string) (string, error)
}

// LocalTmplFactor represents the local template factor
type LocalTmplFactor struct {
	basePath string
}

// NewLocalTmplFactor creates a new local template factor
func NewLocalTmplFactor(basePath string) *LocalTmplFactor {
	return &LocalTmplFactor{
		basePath: basePath,
	}
}

// TmplFactorize returns the factorized template
func (l LocalTmplFactor) TmplFactorize(ctx context.Context, path string) (string, error) {
	filepath := fmt.Sprintf("%s/%s", l.basePath, path)

	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if _, err := buffer.WriteString(scanner.Text()); err != nil {
			return "", fmt.Errorf("failed to write buffer: %v", err)
		}
		if _, err := buffer.WriteString("\n"); err != nil {
			return "", fmt.Errorf("failed to write buffer: %v", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return buffer.String(), nil
}

var _ TmplFactor = (*LocalTmplFactor)(nil)
