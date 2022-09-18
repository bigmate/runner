package runner

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
)

//Runnable is the application interface, an object implementing this interface
//should follow the context protocol
type Runnable interface {
	Run(ctx context.Context) error
}

//Runner is a parent app for apps to be run
type Runner struct {
	apps []Runnable
}

//NewRunner returns *Runner
func NewRunner(apps ...Runnable) *Runner {
	return &Runner{apps: apps}
}

// Add adds the Runnable into collection of Apps to be run and is not thread safe
func (r *Runner) Add(app Runnable) *Runner {
	r.apps = append(r.apps, app)
	return r
}

//Run runs apps passed to the constructor concurrently,
//if one of them fails all the other running apps will be terminated
//by canceling context passed into their Run method
func (r *Runner) Run(ctx context.Context) error {
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	multiErrs := &multierror.Error{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(len(r.apps))

	for _, a := range r.apps {
		go func(a Runnable) {
			defer wg.Done()
			if err := a.Run(ctx); err != nil {
				mu.Lock()
				multiErrs = multierror.Append(multiErrs, err)
				cancel()
				mu.Unlock()
			}
		}(a)
	}

	wg.Wait()

	return multiErrs.ErrorOrNil()
}
