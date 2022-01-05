package state

type GlobalState struct {
	WorldState *WorldState
}

func NewGlobalState() *GlobalState {
	return &GlobalState{
		WorldState: NewWordState(),
	}
}
