package state

import "go-mythril/laser/smt/z3"

// No of iterations to perform when iteration size is symbolic
const APPROX_ITR = 100

type Memory struct {
	msize     int
	rawMemory *map[*z3.Bitvec]*z3.Bitvec
}

func NewMemory() *Memory {
	rawM := make(map[*z3.Bitvec]*z3.Bitvec)
	return &Memory{
		msize:     0,
		rawMemory: &rawM,
	}
}

func (m *Memory) length() int {
	return m.msize
}

func (m *Memory) Extend(size int) {
	m.msize += size
}

func (m *Memory) GetWordAt(index *z3.Bitvec) *z3.Bitvec {
	//TODO:
	return nil
}

func (m *Memory) WriteWordAt(index int, value *z3.Bitvec) {
	//TODO:
}
