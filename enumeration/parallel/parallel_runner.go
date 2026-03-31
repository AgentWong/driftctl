// Package parallel provides a concurrency-limited parallel task runner.
package parallel

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"go.uber.org/atomic"

	"golang.org/x/sync/semaphore"
)

// Runner executes tasks concurrently up to a configurable limit using a semaphore.
type Runner struct {
	sem     *semaphore.Weighted
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	resChan chan interface{}
	err     error
	hasErr  *atomic.Bool
	waiting *atomic.Bool
}

// NewRunner creates a new Runner that allows at most maxRun concurrent tasks.
func NewRunner(ctx context.Context, maxRun int64) *Runner {
	ctx, cancelFunc := context.WithCancel(ctx)
	return &Runner{
		sem:     semaphore.NewWeighted(maxRun),
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancelFunc,
		resChan: make(chan interface{}),
		err:     nil,
		hasErr:  atomic.NewBool(false),
		waiting: atomic.NewBool(false),
	}
}

// SubRunner creates a child Runner that shares the parent's semaphore.
func (p *Runner) SubRunner() *Runner {
	ctx, cancelFunc := context.WithCancel(p.ctx)
	return &Runner{
		sem:     p.sem,
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancelFunc,
		resChan: make(chan interface{}),
		err:     nil,
		hasErr:  atomic.NewBool(false),
		waiting: atomic.NewBool(false),
	}
}

func (p *Runner) Read() chan interface{} {
	p.wait()
	return p.resChan
}

// DoneChan returns a channel that is closed when the runner's context is cancelled.
func (p *Runner) DoneChan() <-chan struct{} {
	return p.ctx.Done()
}

// Err returns the first error encountered by a running task, or nil.
func (p *Runner) Err() error {
	return p.err
}

func (p *Runner) wait() {
	if !p.waiting.Swap(true) {
		go func() {
			p.wg.Wait()
			close(p.resChan)
		}()
	}
}

// Run submits a runnable task to the runner to be executed concurrently.
func (p *Runner) Run(runnable func() (interface{}, error)) {
	p.wg.Add(1)
	go func() {
		if err := p.sem.Acquire(p.ctx, 1); err == nil {
			// only release if sem was acquired
			defer p.sem.Release(1)
		}
		defer p.wg.Done()
		// Prevent new routines executions if we already got an error from another routine
		if p.ctx.Err() != nil {
			return
		}
		// Handle panic in routines and stop runner with proper error
		// Some failed call to grpc plugin like getSchema trigger a panic
		defer func() {
			if r := recover(); r != nil {
				p.Stop(errors.Errorf("A runner routine paniced: %s", r))
			}
		}()
		res, err := runnable()
		if err != nil {
			p.Stop(err)
		}
		p.resChan <- res
	}()
}

// Stop cancels the runner's context with the given error, preventing new tasks from running.
func (p *Runner) Stop(err error) {
	if !p.hasErr.Swap(true) {
		logrus.Debug("Stopping Runner")
		p.err = err
		p.cancel()
	}
}
