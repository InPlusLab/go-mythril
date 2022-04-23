package modules

import (
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
)

type DetectionModule interface {
	ResetModule()
	Execute(target *state.GlobalState) []*analysis.Issue
	_execute(target *state.GlobalState) []*analysis.Issue
}
