package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateFileWithDir creates a file with the directory
func CreateFileWithDir(filePath string) (*os.File, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return f, nil
}
