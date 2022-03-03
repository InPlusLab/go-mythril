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
	opcodes[0x02] = OpcodeTuple{"MUL", 2, 1, 5}
	opcodes[0x03] = OpcodeTuple{"SUB", 2, 1, 3}
	opcodes[0x04] = OpcodeTuple{"DIV", 2, 1, 5}
	opcodes[0x05] = OpcodeTuple{"SDIV", 2, 1, 5}
	opcodes[0x06] = OpcodeTuple{"MOD", 2, 1, 5}
	opcodes[0x07] = OpcodeTuple{"SMOD", 2, 1, 5}
	opcodes[0x08] = OpcodeTuple{"ADDMOD", 2, 1, 8}
	opcodes[0x09] = OpcodeTuple{"MULMOD", 3, 1, 8}
	opcodes[0x0A] = OpcodeTuple{"EXP", 2, 1, 10}
	opcodes[0x0B] = OpcodeTuple{"SIGNEXTEND", 2, 1, 5}

	opcodes[0x10] = OpcodeTuple{"LT", 2, 1, 3}
	opcodes[0x11] = OpcodeTuple{"GT", 2, 1, 3}
	opcodes[0x12] = OpcodeTuple{"SLT", 2, 1, 3}
	opcodes[0x13] = OpcodeTuple{"SGT", 2, 1, 3}
	opcodes[0x14] = OpcodeTuple{"EQ", 2, 1, 3}
	opcodes[0x15] = OpcodeTuple{"ISZERO", 1, 1, 3}
	opcodes[0x16] = OpcodeTuple{"AND", 2, 1, 3}
	opcodes[0x17] = OpcodeTuple{"OR", 2, 1, 3}
	opcodes[0x18] = OpcodeTuple{"XOR", 2, 1, 3}
	opcodes[0x19] = OpcodeTuple{"NOT", 1, 1, 3}
	opcodes[0x1A] = OpcodeTuple{"BYTE", 2, 1, 3}
	opcodes[0x1B] = OpcodeTuple{"SHL", 2, 1, 3}
	opcodes[0x1C] = OpcodeTuple{"SHR", 2, 1, 3}
	opcodes[0x1D] = OpcodeTuple{"SAR", 2, 1, 3}
	opcodes[0x20] = OpcodeTuple{"SHA3", 2, 1, 30}

	opcodes[0x30] = OpcodeTuple{"ADDRESS", 0, 1, 2}
	opcodes[0x31] = OpcodeTuple{"BALANCE", 1, 1, 700}
	opcodes[0x32] = OpcodeTuple{"ORIGIN", 0, 1, 2}
	opcodes[0x33] = OpcodeTuple{"CALLER", 0, 1, 2}
	opcodes[0x34] = OpcodeTuple{"CALLVALUE", 0, 1, 2}
	opcodes[0x35] = OpcodeTuple{"CALLDATALOAD", 1, 1, 3}
	opcodes[0x36] = OpcodeTuple{"CALLDATASIZE", 0, 1, 2}
	opcodes[0x37] = OpcodeTuple{"CALLDATACOPY", 3, 0, 2}
	opcodes[0x38] = OpcodeTuple{"CALLDATASIZE", 0, 1, 2}
	opcodes[0x39] = OpcodeTuple{"CALLDATASIZE", 0, 1, 2}
	opcodes[0x3A] = OpcodeTuple{"CALLDATASIZE", 0, 1, 2}
	opcodes[0x3B] = OpcodeTuple{"CALLDATASIZE", 0, 1, 2}
	opcodes[0x3C] = OpcodeTuple{"EXTCODECOPY", 4, 0, 700}
	opcodes[0x3F] = OpcodeTuple{"EXTCODEHASH", 1, 1, 700}
	opcodes[0x3D] = OpcodeTuple{"RETURNDATASIZE", 0, 1, 2}
	opcodes[0x3E] = OpcodeTuple{"RETURNDATACOPY", 3, 0, 3}

	opcodes[0x40] = OpcodeTuple{"BLOCKHASH", 1, 1, 20}
	opcodes[0x41] = OpcodeTuple{"COINBASE", 0, 1, 2}
	opcodes[0x42] = OpcodeTuple{"TIMESTAMP", 0, 1, 2}
	opcodes[0x43] = OpcodeTuple{"NUMBER", 0, 1, 2}
	opcodes[0x44] = OpcodeTuple{"DIFFICULTY", 0, 1, 2}
	opcodes[0x45] = OpcodeTuple{"GASLIMIT", 0, 1, 2}
	opcodes[0x46] = OpcodeTuple{"CHAINID", 0, 1, 2}
	opcodes[0x47] = OpcodeTuple{"SELFBALANCE", 0, 1, 2}
	opcodes[0x48] = OpcodeTuple{"BASEFEE", 0, 1, 2}

	opcodes[0x50] = OpcodeTuple{"POP", 1, 0, 2}
	opcodes[0x51] = OpcodeTuple{"MLOAD", 1, 1, 3}
	opcodes[0x52] = OpcodeTuple{"MSTORE", 2, 0, 3}
	opcodes[0x53] = OpcodeTuple{"MSTORE8", 2, 0, 3}
	opcodes[0x54] = OpcodeTuple{"SLOAD", 1, 1, 800}
	opcodes[0x55] = OpcodeTuple{"SSTORE", 1, 0, 5000}
	opcodes[0x56] = OpcodeTuple{"JUMP", 1, 0, 8}
	opcodes[0x57] = OpcodeTuple{"JUMPI", 2, 0, 10}
	opcodes[0x58] = OpcodeTuple{"PC", 0, 1, 2}
	opcodes[0x59] = OpcodeTuple{"MSIZE", 0, 1, 2}
	opcodes[0x5A] = OpcodeTuple{"GAS", 0, 1, 2}
	opcodes[0x5B] = OpcodeTuple{"JUMPDEST", 0, 0, 1}
	opcodes[0x5C] = OpcodeTuple{"BEGINSUB", 0, 0, 2}
	opcodes[0x5D] = OpcodeTuple{"RETURNSUB", 0, 0, 5}
	opcodes[0x5E] = OpcodeTuple{"JUMPSUB", 1, 0, 10}

	opcodes[0xA0] = OpcodeTuple{"LOG0", 2, 0, 375}
	opcodes[0xA1] = OpcodeTuple{"LOG1", 3, 0, 750}
	opcodes[0xA2] = OpcodeTuple{"LOG2", 4, 0, 1125}
	opcodes[0xA3] = OpcodeTuple{"LOG3", 5, 0, 1500}
	opcodes[0xA4] = OpcodeTuple{"LOG4", 6, 0, 1875}

	opcodes[0xF0] = OpcodeTuple{"CREATE", 3, 1, 32000}
	opcodes[0xF5] = OpcodeTuple{"CREATE2", 4, 1, 32000}
	opcodes[0xF1] = OpcodeTuple{"CALL", 7, 1, 700}
	opcodes[0xF2] = OpcodeTuple{"CALLCODE", 7, 1, 700}
	opcodes[0xF3] = OpcodeTuple{"RETURN", 2, 0, 0}
	opcodes[0xF4] = OpcodeTuple{"DELEGATECALL", 6, 1, 700}
	opcodes[0xFA] = OpcodeTuple{"STATICCALL", 6, 1, 700}
	opcodes[0xFD] = OpcodeTuple{"REVERT", 2, 0, 0}
	opcodes[0xFF] = OpcodeTuple{"SELFDESTRUCT", 1, 0, 5000}
	opcodes[0xFE] = OpcodeTuple{"INVALID", 0, 0, 0}

	for i := 1; i < 33; i++ {
		opcodes[0x5F+i] = OpcodeTuple{"PUSH" + strconv.Itoa(i), 0, 1, 3}
	}
	for i := 1; i < 17; i++ {
		opcodes[0x7F+i] = OpcodeTuple{"DUP" + strconv.Itoa(i), 0, 0, 0}
	}
	for i := 1; i < 17; i++ {
		opcodes[0x8F+i] = OpcodeTuple{"SWAP" + strconv.Itoa(i), 0, 0, 0}
	}

	return &opcodes
}
