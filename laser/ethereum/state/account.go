package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
)

type Account struct {
	Address  *z3.Bitvec
	Balances *z3.Array
	Code     *disassembler.Disasembly
}

func NewAccount(addr *z3.Bitvec, balances *z3.Array, code *disassembler.Disasembly) *Account {
	return &Account{
		Address:  addr,
		Balances: balances,
		Code:     code,
	}
}

func (acc *Account) Balance() *z3.Bitvec {
	return acc.Balances.GetItem(acc.Address)
}
