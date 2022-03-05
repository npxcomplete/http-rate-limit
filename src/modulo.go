package ratelimit

import "fmt"

type ModuloRel struct {
	Value int
	Mod   int
}

func (modVal ModuloRel) Increment() ModuloRel {
	return modVal.Add(1)
}

func (modVal ModuloRel) Add(x int) ModuloRel {
	return ModuloRel{
		Value: (modVal.Value + x) % modVal.Mod,
		Mod:   modVal.Mod,
	}
}

func (modVal ModuloRel) String() string {
	return fmt.Sprintf("%d `Mod` %d", modVal.Value, modVal.Mod)
}
