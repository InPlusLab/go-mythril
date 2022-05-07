package support

type Args struct {
	SolverTimeout        int
	SparsePruning        bool
	UnconstrainedStorage bool
	ParallelSolving      bool
	CallDepthLimit       int
	Iprof                bool
	SolverLog            string
}

func NewArgs() *Args {
	return &Args{
		SolverTimeout:        10000,
		SparsePruning:        true,
		UnconstrainedStorage: false,
		ParallelSolving:      false,
		CallDepthLimit:       3,
		Iprof:                true,
		SolverLog:            "",
	}
}
