package utils

import "sync"

type CounterWindow struct {
	Mu         sync.Mutex
	LastWindow int64
	CurrWindow int64
	PrevCount  int
	CurrCount  int
}
