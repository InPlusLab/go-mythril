package disassembler

import (
	"encoding/hex"
	"fmt"
)

type Disasembly struct {
	Bytecode        []byte
	InstructionList []*EvmInstruction
}

func NewDisasembly(code string) *Disasembly {
	// TODO: bytecode not creation code
	// 6060320032 just for test
	fmt.Println("NewDisasembly", code)
	if code == "6060320032" {
		bytecode, err := hex.DecodeString(code)
		fmt.Println("bytecode", bytecode, err)
		return &Disasembly{
			Bytecode:        bytecode,
			InstructionList: disassemble(bytecode),
		}
	}

	return nil

}
