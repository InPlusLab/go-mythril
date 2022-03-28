package disassembler

import (
	"encoding/hex"
)

type Disasembly struct {
	Bytecode        []byte
	InstructionList []*EvmInstruction
}

func NewDisasembly(code string) *Disasembly {
	// TODO: bytecode not creation code
	// 6060320032 just for test

	bytecode, _ := hex.DecodeString(code)
	return &Disasembly{
		Bytecode:        bytecode,
		InstructionList: disassemble(bytecode),
	}
}

// TODO
func (d *Disasembly) AssignBytecode(bytecode []byte) {
	d.Bytecode = bytecode
	//d.InstructionList = disassemble(bytecode)
}
