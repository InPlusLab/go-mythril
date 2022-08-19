package modules

import (
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
)

type DetectionModule interface {
	ResetModule()
	Execute(target *state.GlobalState) []*analysis.Issue
	AddIssue(issue *analysis.Issue)
	GetIssues() []*analysis.Issue
	GetPreHooks() []string
	GetPostHooks() []string
	_execute(target *state.GlobalState) []*analysis.Issue
}
