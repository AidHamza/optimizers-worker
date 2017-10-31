package pool

import (
	"github.com/kamilsk/semaphore"
)

type Pool struct {
	sem  semaphore.Semaphore
	work chan func()
}

func (p *Pool) Schedule(task func()) {
	select {
	case p.work <- task: // delay the task to already running workers
	case release, ok := <-p.sem.Signal(nil): if ok { go p.worker(task, release) } // ok is always true in this case
	}
}

func (p *Pool) worker(task func(), release semaphore.ReleaseFunc) {
	defer release()
	var ok bool
	for {
		task()
		task, ok = <-p.work
		if !ok { return }
	}
}

func New(size int) *Pool {
	return &Pool{
		sem:  semaphore.New(size),
		work: make(chan func()),
	}
}