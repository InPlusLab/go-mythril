package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"reflect"
	"sync"
)

// The implementation should be in /module
// Because of the golang cycle import
type PotentialIssue struct {
	Contract        string
	FunctionName    string
	Address         int
	SWCID           string
	Title           string
	Bytecode        []byte
	Severity        string
	DescriptionHead string
	DescriptionTail string
	Constraints     *state.Constraints
	Detector        DetectionModule
}

func (issue *PotentialIssue) Translate(ctx *z3.Context) *PotentialIssue {
	issue.Constraints = issue.Constraints.Translate(ctx)
	return issue
}
func (issue *PotentialIssue) Copy() *PotentialIssue {
	return &PotentialIssue{
		Contract:        issue.Contract,
		FunctionName:    issue.FunctionName,
		Address:         issue.Address,
		SWCID:           issue.SWCID,
		Title:           issue.Title,
		Bytecode:        issue.Bytecode,
		Severity:        issue.Severity,
		DescriptionHead: issue.DescriptionHead,
		DescriptionTail: issue.DescriptionTail,
		Constraints:     issue.Constraints.DeepCopy(),
		Detector:        issue.Detector,
	}
}

type PotentialIssuesAnnotation struct {
	sync.RWMutex
	PotentialIssues []*PotentialIssue
}

func NewPotentialIssuesAnnotation() *PotentialIssuesAnnotation {
	return &PotentialIssuesAnnotation{
		PotentialIssues: make([]*PotentialIssue, 0),
	}
}
func (anno *PotentialIssuesAnnotation) PersistToWorldState() bool {
	return false
}
func (anno *PotentialIssuesAnnotation) PersistOverCalls() bool {
	return false
}
func (anno *PotentialIssuesAnnotation) Copy() state.StateAnnotation {
	newPotentialIssues := make([]*PotentialIssue, 0)
	for _, issue := range anno.Elements() {
		newPotentialIssues = append(newPotentialIssues, issue.Copy())
	}
	newStateAnnotation := &PotentialIssuesAnnotation{
		PotentialIssues: newPotentialIssues,
	}
	return newStateAnnotation
}
func (anno *PotentialIssuesAnnotation) Translate(ctx *z3.Context) state.StateAnnotation {
	newPotentialIssues := make([]*PotentialIssue, 0)
	for _, issue := range anno.Elements() {
		newPotentialIssues = append(newPotentialIssues, issue.Translate(ctx))
	}
	newStateAnnotation := &PotentialIssuesAnnotation{
		PotentialIssues: newPotentialIssues,
	}
	return newStateAnnotation
}
func (anno *PotentialIssuesAnnotation) Append(item ...*PotentialIssue) {
	anno.Lock()
	defer anno.Unlock()
	anno.PotentialIssues = append(anno.PotentialIssues, item...)
}
func (anno *PotentialIssuesAnnotation) Elements() []*PotentialIssue {
	anno.RLock()
	defer anno.RUnlock()
	return anno.PotentialIssues
}
func (anno PotentialIssuesAnnotation) Replace(arr []*PotentialIssue) {
	anno.Lock()
	defer anno.Unlock()
	anno.PotentialIssues = arr
}

func GetPotentialIssuesAnnotaion(globalState *state.GlobalState) *PotentialIssuesAnnotation {
	for _, annotation := range globalState.Annotations {
		if reflect.TypeOf(annotation).String() == "*modules.PotentialIssuesAnnotation" {
			annotation.(*PotentialIssuesAnnotation).RLock()
			defer annotation.(*PotentialIssuesAnnotation).RUnlock()
			return annotation.(*PotentialIssuesAnnotation)
		}
	}
	annotation := NewPotentialIssuesAnnotation()
	globalState.Annotate(annotation)
	return annotation
}

func CheckPotentialIssues(globalState *state.GlobalState) {
	/*
			Called at the end of a transaction, checks potential issues, and
		    adds valid issues to the detector.
	*/
	annotation := GetPotentialIssuesAnnotaion(globalState)
	unsatPotentialIssues := make([]*PotentialIssue, 0)
	for _, potentialIssue := range annotation.Elements() {
		fmt.Println("addPotentialIssues")
		//tmpConstraint := globalState.WorldState.Constraints.Copy()
		tmpConstraint := globalState.WorldState.Constraints.DeepCopy()
		//tmpConstraint.Add(potentialIssue.Constraints.ConstraintList...)
		// should translate the context here
		for _, con := range potentialIssue.Constraints.ConstraintList {
			tmpConstraint.Add(con.Translate(globalState.Z3ctx))
			//tmpConstraint.Add(con)
		}

		//if potentialIssue.Address == 421 ||  potentialIssue.Address == 497 {
		//	fmt.Println(potentialIssue.Address, ":")
		//	for i, con := range tmpConstraint.ConstraintList {
		//		fmt.Println(i, "-", con.BoolString())
		//	}
		//}

		transactionSequence := analysis.GetTransactionSequence(globalState, tmpConstraint)
		if transactionSequence == nil {
			fmt.Println("unsatError")
			// UnsatError
			unsatPotentialIssues = append(unsatPotentialIssues, potentialIssue)
			continue
		}

		// de-duplication: because CheckPotentialIssues() will be triggered more than one time.
		if potentialIssue.Detector.GetCache().Contains(potentialIssue.Address) {
			continue
		} else {
			issue := &analysis.Issue{
				Contract:            potentialIssue.Contract,
				FunctionName:        potentialIssue.FunctionName,
				Address:             potentialIssue.Address,
				Title:               potentialIssue.Title,
				Bytecode:            potentialIssue.Bytecode,
				SWCID:               potentialIssue.SWCID,
				GasUsed:             []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
				Severity:            potentialIssue.Severity,
				DescriptionHead:     potentialIssue.DescriptionHead,
				DescriptionTail:     potentialIssue.DescriptionTail,
				TransactionSequence: transactionSequence,
			}
			potentialIssue.Detector.AddIssue(issue)

			potentialIssue.Detector.GetCache().Add(potentialIssue.Address)
		}

	}
	//annotation.PotentialIssues = unsatPotentialIssues
	annotation.Replace(unsatPotentialIssues)
}
