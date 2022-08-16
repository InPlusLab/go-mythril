package support

type Args struct {
	SolverTimeout        int
	SparsePruning        bool
	UnconstrainedStorage bool
	ParallelSolving      bool
	CallDepthLimit       int
	Iprof                bool
	SolverLog            string
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
		UseIntegerModule:     true,
	}
}
