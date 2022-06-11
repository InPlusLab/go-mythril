package module

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"reflect"
)
// The implementation of security is in /analysis, now it is in /analysis/module
// Because of the circle import problem in Golang.
// TODO:
func RetrieveCallbackIssues(whiteList []string) []*analysis.Issue {
	issues := make([]*analysis.Issue, 0)

	return issues
}

func FireLasers(statespace *state.GlobalState, whiteList []string) []*analysis.Issue {
	fmt.Println("Starting analysis")
	issues := make([]*analysis.Issue, 0)
	// TODO: EntryPoint
	modules := NewModuleLoader().GetDetectionModules(whiteList)
	for _, module := range modules {
		fmt.Println("Executing " + reflect.TypeOf(module).String())
		issues = append(issues, module.Execute(statespace)...)
	}
	// TODO: RetrieveCallbackIssues()
	return issues
}

