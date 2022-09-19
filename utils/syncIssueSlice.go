package utils

import (
	"sync"
)

type SyncIssueSlice struct {
	sync.RWMutex
	arr []interface{}
}

func NewSyncIssueSlice() *SyncIssueSlice {
	return &SyncIssueSlice{
		arr: make([]interface{}, 0),
	}
}

func (s *SyncIssueSlice) Append(item interface{}) {
	s.Lock()
	defer s.Unlock()
	s.arr = append(s.arr, item)
}

func (s *SyncIssueSlice) Load(index int) interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.arr[index]
}

func (s *SyncIssueSlice) Elements() []interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.arr
}

func (s *SyncIssueSlice) Length() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.arr)
}
