package transaction

import (
	"go-mythril/laser/smt/z3"
	"math/big"
)

// In python-mythril, Actors is singleton. But here we think it as a normal class.
type Actors struct {
	Creator   string
	Attacker  string
	Someguy   string
	Addresses map[string]*z3.Bitvec
}

func NewActors(ctx *z3.Context) *Actors {
	creator := "AFFEAFFEAFFEAFFEAFFEAFFEAFFEAFFEAFFEAFFE"
	attacker := "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF"
	//attacker := "5B38Da6a701c568545dCfcB03FcB875f56beddC4"
	someguy := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	creatorBv, _ := new(big.Int).SetString(creator, 16)
	attackerBv, _ := new(big.Int).SetString(attacker, 16)
	someguyBv, _ := new(big.Int).SetString(someguy, 16)

	addrMap := make(map[string]*z3.Bitvec)
	addrMap["CREATOR"] = ctx.NewBitvecVal(creatorBv, 256)
	addrMap["ATTACKER"] = ctx.NewBitvecVal(attackerBv, 256)
	addrMap["SOMEGUY"] = ctx.NewBitvecVal(someguyBv, 256)

	return &Actors{
		Creator:   creator,
		Attacker:  attacker,
		Someguy:   someguy,
		Addresses: addrMap,
	}
}

func (a *Actors) GetCreator() *z3.Bitvec {
	return a.Addresses["CREATOR"]
}
func (a *Actors) GetAttacker() *z3.Bitvec {
	return a.Addresses["ATTACKER"]
}
func (a *Actors) GetSomeGuy() *z3.Bitvec {
	return a.Addresses["SOMEGUY"]
}

// the implementation is in sym.go because of circle import in golang.
//func ExecuteMessageCall(evm *ethereum.LaserEVM, creationCode string, contractName string, inputStr string, ctx *z3.Context) {
//	tx := state.NewMessageCallTransaction(creationCode, contractName, inputStr, ctx)
//	setupGlobalStateForExecution(evm, tx)
//}
//
//func setupGlobalStateForExecution(evm *ethereum.LaserEVM, tx state.BaseTransaction) {
//	globalState := tx.InitialGlobalState()
//	evm.WorkList <- globalState
//}
