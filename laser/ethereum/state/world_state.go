package state

import (
	"fmt"
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
)

type WorldState struct {
	// the key of Accounts is decimal
	Accounts            map[string]*Account
	Balances            *z3.Array
	StartingBalances    *z3.Array
	Constraints         *Constraints
	TransactionSequence []BaseTransaction
}

func NewWordState(ctx *z3.Context) *WorldState {
	//ACTORS := transaction.NewActors(ctx)
	accounts := make(map[string]*Account)
	//attackerAcc := NewAccount(ACTORS.GetAttacker(), nil, false, nil, "")
	//creatorAcc := NewAccount(ACTORS.GetCreator(), nil, false, nil, "")
	//someGuyAcc := NewAccount(ACTORS.GetSomeGuy(), nil, false, nil, "")
	// TODO: just test for balance_()
	//caller, _ := new(big.Int).SetString("5B38Da6a701c568545dCfcB03FcB875f56beddC4", 16)
	balances := ctx.NewArray("balance", 256, 256)
	//balances.SetItem(ACTORS.GetCreator(), ctx.NewBitvec("initBalance",256))
	//balances.SetItem(ACTORS.GetAttacker(), ctx.NewBitvec("initBalance",256))
	//balances.SetItem(ACTORS.GetSomeGuy(), ctx.NewBitvec("initBalance",256))

	//startingBalances := balances.DeepCopy()
	startingBalances := ctx.NewArray("startingBalance", 256, 256)

	ws := &WorldState{
		Accounts:            accounts,
		Balances:            balances,
		StartingBalances:    startingBalances,
		Constraints:         NewConstraints(),
		TransactionSequence: make([]BaseTransaction, 0),
	}

	//ws.PutAccount(attackerAcc)
	//ws.PutAccount(creatorAcc)
	//ws.PutAccount(someGuyAcc)

	return ws
}

func (ws *WorldState) Copy() *WorldState {

	txSeq := make([]BaseTransaction, 0)
	for _, tx := range ws.TransactionSequence {
		txSeq = append(txSeq, tx)
	}
	resWs := &WorldState{
		Accounts:            make(map[string]*Account),
		Balances:            ws.Balances.DeepCopy().(*z3.Array),
		StartingBalances:    ws.StartingBalances.DeepCopy().(*z3.Array),
		Constraints:         ws.Constraints.DeepCopy(),
		TransactionSequence: txSeq,
	}
	for _, acc := range ws.Accounts {
		resWs.PutAccount(acc.Copy())
	}

	return resWs
}

func (ws *WorldState) AccountsExistOrLoad(addr *z3.Bitvec) *Account {
	accounts := ws.Accounts
	acc, ok := accounts[addr.Simplify().Value()]
	if ok {
		return acc
	} else {
		// TODO: find in dynamicLoader
		fmt.Println("don't getAccount in ws!")
		return NewAccount(addr, ws.Balances, false, disassembler.NewDisasembly(""), "")
	}
}

func (ws *WorldState) PutAccount(acc *Account) {
	accounts := ws.Accounts
	accounts[acc.Address.Value()] = acc
	acc.Balances = ws.Balances
}

func (ws *WorldState) CreateAccount(balance int, concreteStorage bool, creator *z3.Bitvec, address *z3.Bitvec, code *disassembler.Disasembly, contractName string) *Account {
	var trueAddr *z3.Bitvec
	ctx := creator.GetCtx()
	if address == nil {
		// TODO: _generate_new_address(creator)
		trueAddr = creator.BvAdd(ctx.NewBitvecVal(1, 256)).Simplify()
	} else {
		trueAddr = address
	}
	newAccount := NewAccount(trueAddr, ws.Balances, concreteStorage, code, contractName)

	newAccount.SetBalance(ctx.NewBitvecVal(balance, 256))
	ws.PutAccount(newAccount)

	return newAccount
}

func (ws *WorldState) Translate(ctx *z3.Context) *WorldState {
	newConstraints := NewConstraints()
	for _, v := range ws.Constraints.ConstraintList {
		newV := v.Translate(ctx)
		newConstraints.Add(newV)
	}
	newAccouts := make(map[string]*Account)
	for i, v := range ws.Accounts {
		newAccouts[i] = v.Translate(ctx)
	}

	return &WorldState{
		Accounts:            newAccouts,
		Balances:            ws.Balances.Translate(ctx).(*z3.Array),
		StartingBalances:    ws.StartingBalances.Translate(ctx).(*z3.Array),
		Constraints:         newConstraints,
		TransactionSequence: ws.TransactionSequence,
	}
}
