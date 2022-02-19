package support

import "strconv"

type OpcodeTuple struct {
	Name   string
	Stack0 int
	Stack1 int
	Gas    int
}

//TODO
func NewOpcodes() *map[int]OpcodeTuple {
	opcodes := make(map[int]OpcodeTuple)
	opcodes[0x00] = OpcodeTuple{"STOP", 0, 0, 0}
	opcodes[0x01] = OpcodeTuple{"ADD", 2, 1, 3}
	opcodes[0x32] = OpcodeTuple{"ORIGIN", 0, 1, 2}
	opcodes[0x42] = OpcodeTuple{"TIMESTAMP", 0, 1, 2}

	for i := 1; i < 33; i++ {
		opcodes[0x5F+i] = OpcodeTuple{"PUSH" + strconv.Itoa(i), 0, 1, 3}
	}

	return &opcodes
}
