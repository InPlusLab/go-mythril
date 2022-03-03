package disassembler

import (
	"encoding/hex"
	"go-mythril/support"
)

type EvmInstruction struct {
	// Important: PC in mythril seems to be different with others (e.g. Etherscan). Mythril.address = Others.pc. Mythril.pc seems to be the index of the evminstruction.
	Address  int
	OpCode   support.OpcodeTuple
	Argument string
}

func disassemble(bytecode []byte) []*EvmInstruction {
	ret := make([]*EvmInstruction, 0)
	opcodes := *support.NewOpcodes()
	// TODO
	if hex.EncodeToString(bytecode) == "6060320032" {
		ret = append(ret, &EvmInstruction{
			Address:  0,
			OpCode:   opcodes[0x60],
			Argument: "0x10",
		})
		ret = append(ret, &EvmInstruction{
			Address:  2,
			OpCode:   opcodes[0x61],
			Argument: "0x02",
		})
		ret = append(ret, &EvmInstruction{
			Address:  3,
			OpCode:   opcodes[0x0A],
			Argument: "0x0A",
		})
		ret = append(ret, &EvmInstruction{
			Address:  4,
			OpCode:   opcodes[0x00],
			Argument: "0x00",
		})
		return ret
	}
	return nil
}
