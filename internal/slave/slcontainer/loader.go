package slcontainer

import (
	"fmt"
	"strings"
	"sync"
)

// Loader represents the loader container for the slave node
type Loader struct {
	mu               *sync.RWMutex
	LoaderBuilderMap map[string]*strings.Builder
	LoaderMap        map[string]string
}

// NewLoader creates a new loader container for the slave node
func NewLoader() *Loader {
	return &Loader{
		mu:               &sync.RWMutex{},
		LoaderBuilderMap: make(map[string]*strings.Builder),
		LoaderMap:        make(map[string]string),
	}
}

// WriteString writes a string to the loader container
func (l *Loader) WriteString(loaderID, data string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.LoaderBuilderMap[loaderID]; !ok {
		l.LoaderBuilderMap[loaderID] = &strings.Builder{}
	}
	l.LoaderBuilderMap[loaderID].WriteString(data)
}

// Build builds the loader container
func (l *Loader) Build(loaderID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Println("Building loader", loaderID)

	fmt.Println("LoaderBuilderMap", l.LoaderBuilderMap)

	if _, ok := l.LoaderBuilderMap[loaderID]; ok {
		fmt.Println("Built loader", loaderID)
		l.LoaderMap[loaderID] = l.LoaderBuilderMap[loaderID].String()
		delete(l.LoaderBuilderMap, loaderID)
	}

	fmt.Println("LoaderMap", l.LoaderMap)
}

// GetLoader returns the loader from the container
func (l *Loader) GetLoader(loaderID string) (string, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if loader, ok := l.LoaderMap[loaderID]; ok {
		return loader, true
	}
	return "", false
}

// LoaderResourceRequest is a struct that represents a request to the loader resource.
type LoaderResourceRequest struct {
	LoaderID string
}
