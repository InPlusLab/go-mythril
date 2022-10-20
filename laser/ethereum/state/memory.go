package state

import (
	"encoding/hex"
	"fmt"
	"go-mythril/laser/smt/z3"
	"sort"
	"strconv"
	"strings"
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
func (m *Memory) PrintMemoryOneLine() {
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
		head := keyArr[0]
		for i, v := range keyArr {
			//str := value.BvString()
			//idx := strings.Index(str, "\n")
			//if idx == -1 {
			//	fmt.Println("PrintStack: ", str, " ", m.RawStack[i].Annotations)
			//} else {
			//	fmt.Println("PrintStack: ", str[:idx], " ", m.RawStack[i].Annotations)
			//}

			if v < head+16 {
				value, ok := mem[int64(v)]
				if ok {
					str := value.BvString()
					idx := strings.Index(str, "\n")
					if idx == -1 {
						tmpStr += value.BvString()[2:]
					} else {
						tmpStr += value.BvString()[2:idx]
					}
				}
			} else {
				fmt.Println("PrintMem", strconv.FormatInt(int64(head), 16), ":", tmpStr)
				head = keyArr[i]
				value, ok := mem[int64(v)]
				tmpStr = "0x"
				if ok {
					//tmpStr += value.BvString()[2:]
					str := value.BvString()
					idx := strings.Index(str, "\n")
					if idx == -1 {
						tmpStr += value.BvString()[2:]
					} else {
						tmpStr += value.BvString()[2:idx]
					}
				}
			}
		}
		fmt.Println("PrintMem", strconv.FormatInt(int64(head), 16), ":", tmpStr)
	}
}

func (m *Memory) PrintMemory() {
	mem := *m.RawMemory
	fmt.Println(mem)
	if len(mem) == 0 {
		fmt.Println("PrintMemory: null")
	} else {

		keyArr := make([]int, 0)
		for i, _ := range mem {
			keyArr = append(keyArr, int(i))
		}
		sort.Ints(keyArr)
		tmpStr := "0x"
		head := keyArr[0]
		for i, v := range keyArr {
			if v < head+16 {
				value, ok := mem[int64(v)]
				if ok {
					tmpStr += value.BvString()[2:]
				}
			} else {
				fmt.Println("PrintMem", strconv.FormatInt(int64(head), 16), ":", tmpStr)
				head = keyArr[i]
				value, ok := mem[int64(v)]
				tmpStr = "0x"
				if ok {
					tmpStr += value.BvString()[2:]
				}
			}
		}
		fmt.Println("PrintMem", strconv.FormatInt(int64(head), 16), ":", tmpStr)
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
	result, ok := mem[index]
	fmt.Println("result:", result, "ok:", ok)
	if !ok {
		var ctx *z3.Context
		for _, v := range mem {
			ctx = v.GetCtx()
			break
		}
		return ctx.NewBitvecVal(0, 256)
	} else {
		ctx := result.GetCtx()
		for i := index + 1; i < index+32; i++ {
			if mem[i] != nil {
				result = result.Concat(mem[i])
			} else {
				result = result.Concat(ctx.NewBitvecVal(0, 8))
			}
		}
		result = result.Simplify()
		if result.BvSize() != 256 {
			panic("memory size error in GetWordAt")
		}
		return result
	}

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
	//length := stop - start
	for i := start; i <= stop; i++ {
		value, ok := mem[i]
		if ok {
			items = append(items, value)
		} else {
			fmt.Println("notGetItems")
		}
	}
	return items
}

func (m *Memory) GetItems2Bytes(start int64, stop int64) []byte {
	bvarr := m.GetItems(start, stop)
	str := ""
	fmt.Println("11")
	for _, item := range bvarr {
		// TODO: array getItm symbolic false
		str += item.Value()
	}
	fmt.Println("33")
	bytecode, _ := hex.DecodeString(str)
	fmt.Println("44")
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

func (m *Memory) Copy() *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	for i, v := range *m.RawMemory {
		rawM[i] = v
	}
	return &Memory{
		Msize:     m.Msize,
		RawMemory: &rawM,
	}
}

func (m *Memory) CopyTranslate(ctx *z3.Context) *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	for i, v := range *m.RawMemory {
		rawM[i] = v.Translate(ctx)
	}
	return &Memory{
		Msize:     m.Msize,
		RawMemory: &rawM,
	}
}
