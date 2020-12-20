package main

type Strategy interface {
	Select(alives []bool) (int, error)
}

type RoundRobinStrategy struct {
	current int
}

func (rr *RoundRobinStrategy) Select(alives []bool) (int, error) {
	rr.current++

	for !alives[rr.current%len(alives)] {
		rr.current++
	}
	return rr.current % len(alives), nil
}
