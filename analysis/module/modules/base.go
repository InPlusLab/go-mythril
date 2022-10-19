package modules

import (
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
)

type DetectionModule interface {
	ResetModule()
	Execute(target *state.GlobalState) []*analysis.Issue
	AddIssue(issue *analysis.Issue)
	GetIssues() []*analysis.Issue
	GetPreHooks() []string
	GetPostHooks() []string
	GetCache() *utils.Set
	_execute(target *state.GlobalState) []*analysis.Issue
}
