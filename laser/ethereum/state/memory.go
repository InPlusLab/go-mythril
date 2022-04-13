package state

import (
	"encoding/hex"
	"fmt"
	"go-mythril/laser/smt/z3"
	"sort"
)

// No of iterations to perform when iteration size is symbolic
const APPROX_ITR = 100

type Memory struct {
	Msize int
	// the value of RawMemory is (bv, 8)
	RawMemory *map[int64]*z3.Bitvec
}

func NewMemory() *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	return &Memory{
		Msize:     0,
		RawMemory: &rawM,
	}
}

// For debug
func (m *Memory) PrintMemory() {
	mem := *m.RawMemory
	if len(mem) == 0 {
		fmt.Println("PrintMemory: null")
	} else {
		keyArr := make([]int, 0)
		for i, _ := range mem {
			keyArr = append(keyArr, int(i))
		}
		sort.Ints(keyArr)
		tmpStr := "0x"
		count := 0
		lastIndex := 0
		for i, v := range keyArr {
			value, ok := mem[int64(v)]
			if ok {
				tmpStr += value.String()[2:]
				count++
			}
			if count == 16 {
				fmt.Println("PrintMem:", v-15, "-", tmpStr)
				tmpStr = "0x"
				count = 0
				lastIndex = i
			}
		}
		if lastIndex < len(keyArr) {
			str := "0x"
			for k := lastIndex + 1; k < len(keyArr); k++ {
				str += mem[int64(keyArr[k])].String()[2:]
			}
			fmt.Println("PrintMem:", lastIndex+1, "-", tmpStr)
		}
	}
}

func (m *Memory) length() int {
	return m.Msize
}

func (m *Memory) Extend(size int) {
	m.Msize += size
}

func (m *Memory) GetWordAt(index int64) *z3.Bitvec {
	mem := *m.RawMemory
	result := mem[index]
	for i := index + 1; i < index+32; i++ {
		result = result.Concat(mem[i])
	}
	result = result.Simplify()
	if result.BvSize() != 256 {
		panic("memory size error in GetWordAt")
	}
	return result
}

func (m *Memory) WriteWordAt(index int64, value *z3.Bitvec) {
	if value.BvSize() != 256 {
		panic("memory size error in WriteWordAt")
	}
	mem := *m.RawMemory
	for i := 0; i < 256; i = i + 8 {
		mem[index+31-int64(i/8)] = value.Extract(i+7, i).Simplify()
	}
}

func (m *Memory) GetItem(index int64) *z3.Bitvec {
	mem := *m.RawMemory
	return mem[index]
}

func (m *Memory) GetItems(start int64, stop int64) []*z3.Bitvec {
	items := make([]*z3.Bitvec, 0)
	mem := *m.RawMemory
	length := stop - start
	for i := start; i < length; i++ {
		value, ok := mem[i]
		if ok {
			items = append(items, value)
		}
	}
	return items
}

func (m *Memory) GetItems2Bytes(start int64, stop int64) []byte {
	bvarr := m.GetItems(start, stop)
	str := ""
	for _, item := range bvarr {
		str += item.Value()
	}
	bytecode, _ := hex.DecodeString(str)
	return bytecode
}

func (m *Memory) SetItem(key int64, value *z3.Bitvec) {
	memory := *m.RawMemory
	//if int(key) >= len(memory) {
	//	fmt.Println("lenm")
	//	return
	//}
	if value.BvSize() != 8 {
		fmt.Println("bvsize")
		return
	}

	memory[key] = value
}
