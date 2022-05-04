package analysis

import (
	"go-mythril/laser/ethereum/state"
	"reflect"
)

type PotentialIssue struct {
	Contract     string
	FunctionName string
	Address      int
	SWCID        string
	Title        string
	Bytecode     []byte
	// TODO: detetor
	Severity        string
	DescriptionHead string
	DescriptionTail string
	Constraints     *state.Constraints
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
