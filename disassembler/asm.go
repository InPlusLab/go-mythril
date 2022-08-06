package disassembler

import (
	"encoding/hex"
	"go-mythril/support"
	"strconv"
	"strings"
)

type EvmInstruction struct {
	// Important: PC in mythril seems to be different with others (e.g. Etherscan). Mythril.address = Others.pc. Mythril.pc seems to be the index of the evminstruction.
	Address  int
	OpCode   support.OpcodeTuple
	Argument string
}

func disassemble(bytecode []byte) []*EvmInstruction {
	ret := make([]*EvmInstruction, 0)
	address := 0
	length := len(bytecode)
	//partCode := bytecode[(length - 43):]
	opcodes := *support.NewOpcodes()

	// Ignore swarm hash
	/*	partcodeStr := hex.EncodeToString(partCode)
		if strings.Index(partcodeStr, "bzzr") != -1 {
			length -= 43
		}*/

	for address < length {
		opCode := opcodes[int(bytecode[address])]
		// TODO: opCode Invalid check
		currentInstr := &EvmInstruction{
			Address: address,
			OpCode:  opCode,
		}

		match := strings.HasPrefix(opCode.Name, "PUSH")
		if match {
			value, _ := strconv.ParseInt(opCode.Name[4:], 10, 64)
			argumentBytes := bytecode[address+1 : address+1+int(value)]
			currentInstr.Argument = "0x" + hex.EncodeToString(argumentBytes)
			address += int(value)
		}
		// For debug
		//fmt.Println(currentInstr)
		ret = append(ret, currentInstr)
		address += 1
	}

	return ret

	/*	if hex.EncodeToString(bytecode) == "6060320032" {
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
		return nil*/
}
