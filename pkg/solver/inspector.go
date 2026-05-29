package solver

import (
	"fmt"
	"sync"
)

type Inspector struct {
	step    chan struct{}
	release chan struct{}
	Done    chan struct{}

	OnCheckpoint func()

	mu         sync.Mutex
	iter       int
	paramNames []string
	F          []float64
	J          [][]float64
	paused     bool
	Enabled    bool
}

func NewInspector(enabled bool) *Inspector {
	return &Inspector{
		step:    make(chan struct{}),
		release: make(chan struct{}),
		Done:    make(chan struct{}),
		Enabled: enabled,
	}
}

func (ins *Inspector) Checkpoint(iteration int, paramNames []string, J [][]float64, F []float64) {
	if !ins.Enabled {
		return
	}

	ins.mu.Lock()
	ins.iter = iteration
	ins.paramNames = paramNames
	ins.J = J
	ins.F = F
	ins.mu.Unlock()

	if ins.OnCheckpoint != nil {
		ins.OnCheckpoint()
	}

	ins.mu.Lock()
	ins.paused = true
	ins.mu.Unlock()

	fmt.Printf("[inspector] paused at iteration %d\n", iteration)
	select {
	case <-ins.step:
		fmt.Printf("[inspector] stepping iteration %d\n", iteration)
	case <-ins.release:
		fmt.Println("[inspector] released — running freely")
		ins.mu.Lock()
		ins.Enabled = false
		ins.mu.Unlock()
	}

	ins.mu.Lock()
	ins.paused = false
	ins.mu.Unlock()
}

func (ins *Inspector) Step() {
	ins.mu.Lock()
	p := ins.paused
	ins.mu.Unlock()
	if p {
		ins.step <- struct{}{}
	}
}

func (ins *Inspector) Release() {
	ins.mu.Lock()
	p := ins.paused
	ins.mu.Unlock()
	if p {
		ins.release <- struct{}{}
	}
}

func (ins *Inspector) IsPaused() bool {
	ins.mu.Lock()
	defer ins.mu.Unlock()
	return ins.paused
}

func (ins *Inspector) DisplayData() (iter int, paramNames []string, J [][]float64, F []float64) {
	ins.mu.Lock()
	defer ins.mu.Unlock()
	return ins.iter, ins.paramNames, ins.J, ins.F
}
