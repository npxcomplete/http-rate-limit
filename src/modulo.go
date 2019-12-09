package ratelimit

import "fmt"

type ModuloRel struct {
	value int
	mod   int
}

func (modVal ModuloRel) Increment() ModuloRel {
	return modVal.Add(1)
}

func (modVal ModuloRel) Add(x int) ModuloRel {
	return ModuloRel{
		value: (modVal.value + x) % modVal.mod,
		mod:   modVal.mod,
	}
}

func (modVal ModuloRel) String() string {
	return fmt.Sprintf("%d `mod` %d", modVal.value, modVal.mod)
}
