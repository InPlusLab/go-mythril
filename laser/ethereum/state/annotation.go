package state

type StateAnnotation interface {
	PersistToWorldState() bool
	PersistOverCalls() bool
	Copy() StateAnnotation
}
