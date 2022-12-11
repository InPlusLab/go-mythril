package utils

import (
	"sync"
)

type SyncSlice struct {
	sync.RWMutex
	arr []interface{}
}

func NewSyncSlice() *SyncSlice {
	return &SyncSlice{
		arr: make([]interface{}, 0),
	}
}

func NewSyncSliceWithArr(arr []interface{}) *SyncSlice {
	return &SyncSlice{
		arr: arr,
	}
}

func (s *SyncSlice) Append(item interface{}) {
	s.Lock()
	defer s.Unlock()
	s.arr = append(s.arr, item)
}

func (s *SyncSlice) Load(index int) interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.arr[index]
}

func (s *SyncSlice) Elements() []interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.arr
}

func (s *SyncSlice) Length() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.arr)
}
