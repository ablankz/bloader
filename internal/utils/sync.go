package utils

import "sync"

func NewSyncMapFromMap(m map[string]any) *sync.Map {
	sm := &sync.Map{}
	for k, v := range m {
		sm.Store(k, v)
	}
	return sm
}
