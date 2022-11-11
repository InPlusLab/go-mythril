package strategy

type BoundedLoopsStrategy struct {
	Bound int
}

func NewBoundedLoopsStrategy(bound int) BoundedLoopsStrategy {
	return BoundedLoopsStrategy{
		Bound: bound,
	}
}

func (b BoundedLoopsStrategy) CalculateHash(i int, j int, trace []int) int {
	key := 0
	size := 0
	for itr := i; itr < j; itr++ {
		tmp := trace[itr] << ((itr - i) * 8)
		key = key | tmp
		size = size + 1
	}
	return key
}

func (b BoundedLoopsStrategy) CountKey(trace []int, key int, start int, size int) int {
	count := 1
	i := start

KEYLOOP:
	for {
		if i < 0 {
			break KEYLOOP
		}
		if b.CalculateHash(i, i+size, trace) != key {
			break KEYLOOP
		}
		count = count + 1
		i = i - size
	}
	return count
}

func (b BoundedLoopsStrategy) GetLoopCount(trace []int) int {
	found := false
	length := len(trace)
	var i int
	for i = length - 3; i > 0; i = i - 1 {
		if trace[i] == trace[length-2] && trace[i+1] == trace[length-1] {
			found = true
			break
		}
	}

	var count int
	if found {
		key := b.CalculateHash(i+1, length-1, trace)
		size := length - i - 2
		count = b.CountKey(trace, key, i+1, size)
	} else {
		count = 0
	}
	return count
}
