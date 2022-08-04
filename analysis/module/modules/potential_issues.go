package modules

import (
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"reflect"
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

type PotentialIssuesAnnotation struct {
	PotentialIssues []*PotentialIssue
}

func NewPotentialIssuesAnnotation() PotentialIssuesAnnotation {
	return PotentialIssuesAnnotation{
		PotentialIssues: make([]*PotentialIssue, 0),
	}
}
func (anno PotentialIssuesAnnotation) PersistToWorldState() bool {
	return false
}
func (anno PotentialIssuesAnnotation) PersistOverCalls() bool {
	return false
}

func GetPotentialIssuesAnnotaion(globalState *state.GlobalState) PotentialIssuesAnnotation {
	for _, annotation := range globalState.Annotations {
		if reflect.TypeOf(annotation).String() == "PotentialIssuesAnnotation" {
			return annotation.(PotentialIssuesAnnotation)
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
	for _, potentialIssue := range annotation.PotentialIssues {
		tmpConstraint := globalState.WorldState.Constraints.Copy()
		tmpConstraint.Add(potentialIssue.Constraints.ConstraintList...)
		transactionSequence := analysis.GetTransactionSequence(globalState, tmpConstraint)
		if transactionSequence == nil {
			// UnsatError
			unsatPotentialIssues = append(unsatPotentialIssues, potentialIssue)
			continue
		}
		// TODO: potentialIssue.detetor.cache.add( potentialIssue.address )
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
		issues := potentialIssue.Detector.GetIssues()
		issues = append(issues, issue)
	}
	annotation.PotentialIssues = unsatPotentialIssues
}
