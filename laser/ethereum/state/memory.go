package state

import (
	"encoding/hex"
	"fmt"
	"go-mythril/laser/smt/z3"
	"strconv"
)

// No of iterations to perform when iteration size is symbolic
const APPROX_ITR = 100

/*
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
		panic("memory size error in WriteWordAt" + strconv.Itoa(value.BvSize()))
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
	for _, item := range bvarr {
		// TODO: array getItm symbolic false
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

func (m *Memory) Copy() *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	for i, v := range *m.RawMemory {
		rawM[i] = v.Copy()
	}
	return &Memory{
		Msize:     m.Msize,
		RawMemory: &rawM,
	}
}

func (m *Memory) CopyTranslate(ctx *z3.Context) *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	for i, v := range *m.RawMemory {
		//rawM[i] = v.Translate8(ctx)
		rawM[i] = v.Translate(ctx)
	}
	return &Memory{
		Msize:     m.Msize,
		RawMemory: &rawM,
	}
}
*/

type Memory struct {
	Msize int
	// the value of RawMemory is (bv, 8)
	RawMemory    *map[int64]*z3.Bitvec
	RawSymMemory *map[*z3.Bitvec]*z3.Bitvec
	//RawSymMemory *map[string]*z3.Bitvec
}

func NewMemory() *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	rawSymM := make(map[*z3.Bitvec]*z3.Bitvec)
	return &Memory{
		Msize:        0,
		RawMemory:    &rawM,
		RawSymMemory: &rawSymM,
	}
}

func (m *Memory) SymLoad(key *z3.Bitvec) *z3.Bitvec {
	for k, _ := range *m.RawSymMemory {
		if k.Eq(key).IsTrue() {
			return k
		}
	}
	return nil
}

func (m *Memory) length() int {
	return m.Msize
}

func (m *Memory) Extend(size int) {
	m.Msize += size
}

func (m *Memory) GetWordAt(index *z3.Bitvec) *z3.Bitvec {

	if index.Symbolic() {
		//if index.ValueInt() == 0 && index.BvString() != "#x0000000000000000000000000000000000000000000000000000000000000000" {
		mem := *m.RawSymMemory
		//result, ok := mem[index.Simplify().BvString()]
		//if ok {
		//	return result
		//}else {
		//	fmt.Println("before")
		//	fmt.Println("beforeSim:", index.BvString())
		//	fmt.Println("afterSim:", index.Simplify().BvString())
		//	panic("shit!")
		//}

		//if !ok {
		//	fmt.Println("didn't get obj in memory symbolic")
		//	var ctx *z3.Context
		//	for _, v := range *m.RawMemory {
		//		ctx = v.GetCtx()
		//		break
		//	}
		//	return ctx.NewBitvecVal(0, 256)
		//} else {
		//	ctx := result.GetCtx()
		//	for i := 1; i < 32; i++ {
		//		index = index.BvAdd(ctx.NewBitvecVal(i,256)).Simplify()
		//		if mem[index.BvString()] != nil {
		//			result = result.Concat(mem[index.BvString()])
		//		} else {
		//			result = result.Concat(ctx.NewBitvecVal(1, 8))
		//		}
		//	}
		//	result = result.Simplify()
		//	if result.BvSize() != 256 {
		//		panic("memory size error in GetWordAt")
		//	}
		//	return result
		//}

		key := m.SymLoad(index)
		if key == nil {
			fmt.Println("didn't get obj in memory symbolic")
			var ctx *z3.Context
			for _, v := range *m.RawMemory {
				ctx = v.GetCtx()
				break
			}
			return ctx.NewBitvecVal(0, 256)
		} else {
			result := mem[key]
			//ctx := result.GetCtx()
			//for i := 1; i < 32; i++ {
			//	key = key.BvAdd(ctx.NewBitvecVal(i, 256)).Simplify()
			//	realIndex := m.SymLoad(key)
			//	if realIndex != nil {
			//		result = result.Concat(mem[realIndex])
			//	} else {
			//		result = result.Concat(ctx.NewBitvecVal(0, 8))
			//	}
			//}
			result = result.Simplify()
			if result.BvSize() != 256 {
				panic("memory size error in GetWordAt")
			}
			//if result.RealSymbolic() {
			//	result.SetSymbolic(true)
			//}else{
			//	result.SetSymbolic(false)
			//}
			return result
		}
	} else {
		mem := *m.RawMemory
		indexV, _ := strconv.ParseInt(index.Value(), 10, 64)
		result, ok := mem[indexV]

		if !ok {
			fmt.Println("didn't get obj in memory concrete")
			var ctx *z3.Context
			for _, v := range mem {
				ctx = v.GetCtx()
				break
			}
			return ctx.NewBitvecVal(0, 256)
		} else {
			ctx := result.GetCtx()
			for i := indexV + 1; i < indexV+32; i++ {
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
			//if result.RealSymbolic() {
			//	result.SetSymbolic(true)
			//}else{
			//	result.SetSymbolic(false)
			//}
			return result
		}
	}

}

func (m *Memory) WriteWordAt(index *z3.Bitvec, value *z3.Bitvec) {
	if value.BvSize() != 256 {
		panic("memory size error in WriteWordAt" + strconv.Itoa(value.BvSize()))
	}
	if index.Symbolic() {
		//if index.ValueInt() == 0 && index.BvString() != "#x0000000000000000000000000000000000000000000000000000000000000000" {
		mem := *m.RawSymMemory
		key := m.SymLoad(index)
		if key == nil {
			mem[index] = value
		} else {
			mem[key] = value
		}

		//mem[index.Simplify().BvString()] = value

		//ctx := index.GetCtx()
		//for i := 0; i < 256; i = i + 8 {
		//	//offset := index.BvAdd(ctx.NewBitvecVal(31-(i/8), 256)).Simplify()
		//	offset := index.BvAdd(ctx.NewBitvecVal(31, 256)).BvSub(ctx.NewBitvecVal(i/8, 256)).Simplify()
		//
		//	//fmt.Println(offset)
		//	//key := m.SymLoad(offset)
		//	//if key == nil {
		//	//	mem[offset] = value.Extract(i+7, i).Simplify()
		//	//} else {
		//	//	mem[key] = value.Extract(i+7, i).Simplify()
		//	//}
		//	mem[offset.BvString()] = value.Extract(i+7, i).Simplify()
		//	//mem["index.BvString()" + strconv.Itoa(i)] = value.Extract(i+7, i).Simplify()
		//}
	} else {
		mem := *m.RawMemory
		indexV, _ := strconv.ParseInt(index.Value(), 10, 64)
		for i := 0; i < 256; i = i + 8 {
			mem[indexV+31-int64(i/8)] = value.Extract(i+7, i).Simplify()
		}
	}

}

func (m *Memory) GetItem(index interface{}) *z3.Bitvec {
	switch index.(type) {
	//case string:
	//	mem := *m.RawSymMemory
	//	return mem[index.(string)]
	case *z3.Bitvec:
		mem := *m.RawSymMemory
		key := m.SymLoad(index.(*z3.Bitvec))
		if key != nil {
			return mem[key]
		} else {
			panic("can't getItem")
		}
	case int64:
		mem := *m.RawMemory
		return mem[index.(int64)]
	default:
		mem := *m.RawMemory
		return mem[index.(int64)]
	}
}

func (m *Memory) GetItems(start int64, stop int64, ctx *z3.Context) []*z3.Bitvec {
	items := make([]*z3.Bitvec, 0)
	mem := *m.RawMemory
	//length := stop - start
	for i := start; i <= stop; i++ {
		value, ok := mem[i]
		if ok {
			items = append(items, value)
		} else {
			// fmt.Println("notGetItems")
		}
	}
	return items
}

func (m *Memory) GetItems2Bytes(start int64, stop int64, ctx *z3.Context) []byte {
	bvarr := m.GetItems(start, stop, ctx)
	str := ""
	for _, item := range bvarr {
		// TODO: array getItm symbolic false
		str += item.Value()
	}
	bytecode, _ := hex.DecodeString(str)
	return bytecode
}

func (m *Memory) SetItem(key *z3.Bitvec, value *z3.Bitvec) {

	if value.BvSize() != 8 {
		return
	}
	if key.Symbolic() {
		//memory := *m.RawSymMemory
		//memory[key.BvString()] = value
		memory := *m.RawSymMemory
		realKey := m.SymLoad(key)
		if realKey == nil {
			memory[key] = value
		} else {
			memory[realKey] = value
		}

	} else {
		memory := *m.RawMemory
		keyV, _ := strconv.ParseInt(key.Value(), 10, 64)
		memory[keyV] = value
	}
}

func (m *Memory) Copy() *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	for i, v := range *m.RawMemory {
		rawM[i] = v.Copy()
	}
	rawSymM := make(map[*z3.Bitvec]*z3.Bitvec)
	//rawSymM := make(map[string]*z3.Bitvec)
	for i, v := range *m.RawSymMemory {
		rawSymM[i.Copy()] = v.Copy()
		//rawSymM[i] = v.Copy()
	}
	return &Memory{
		Msize:        m.Msize,
		RawMemory:    &rawM,
		RawSymMemory: &rawSymM,
	}
}

func (m *Memory) CopyTranslate(ctx *z3.Context) *Memory {
	rawM := make(map[int64]*z3.Bitvec)
	for i, v := range *m.RawMemory {
		//rawM[i] = v.Translate8(ctx)
		rawM[i] = v.Translate(ctx)
	}
	rawSymM := make(map[*z3.Bitvec]*z3.Bitvec)
	//rawSymM := make(map[string]*z3.Bitvec)
	for i, v := range *m.RawSymMemory {
		rawSymM[i.Translate(ctx)] = v.Translate(ctx)
		//rawSymM[i] = v.Translate(ctx)
	}
	return &Memory{
		Msize:        m.Msize,
		RawMemory:    &rawM,
		RawSymMemory: &rawSymM,
	}
}
