package support

type Args struct {
	SolverTimeout        int
	SparsePruning        bool
	UnconstrainedStorage bool
	ParallelSolving      bool
	CallDepthLimit       int
	Iprof                bool
	SolverLog            string
	// TxSeq: each tx has a list for transaction Hash strings.
	TransactionSequences [][]string
	UseIntegerModule     bool
}

func NewArgs() *Args {
	return &Args{
		SolverTimeout:        30000,
		SparsePruning:        true,
		UnconstrainedStorage: false,
		ParallelSolving:      false,
		CallDepthLimit:       3,
		Iprof:                true,
		SolverLog:            "",
		TransactionSequences: make([][]string, 0),
		UseIntegerModule:     true,
	}
}
