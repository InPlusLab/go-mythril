package ethereum

import "go-mythril/disassembler"

func GetInstructionIndex(instrList []*disassembler.EvmInstruction, addr int) int {
	index := 0
	for _, instr := range instrList {
		if instr.Address >= addr {
			return index
		}
		index += 1
	}
	return -1
}
