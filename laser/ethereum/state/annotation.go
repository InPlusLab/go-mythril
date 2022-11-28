package state

import "go-mythril/laser/smt/z3"

type StateAnnotation interface {
	PersistToWorldState() bool
	PersistOverCalls() bool
	Copy() StateAnnotation
	Translate(ctx *z3.Context) StateAnnotation
}
