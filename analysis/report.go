package analysis

type Issue struct {
	Contract        string
	FunctionName    string
	Address         int
	SWCID           string
	Title           string
	Bytecode        []byte
	Severity        string
	DescriptionHead string
	DescriptionTail string
	GasUsed         []int
	SourceLocation  int
}

func NewIssue(contract string, funcName string, addr int, swcId string,
	title string, bytecode []byte, severity string) *Issue {
	return &Issue{
		Contract:     contract,
		FunctionName: funcName,
		Address:      addr,
		SWCID:        swcId,
		Title:        title,
		Bytecode:     bytecode,
		Severity:     severity,
	}
}
