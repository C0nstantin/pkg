// Package closer provides a simple way to close all resources in a graceful way
// based on os.Signal

package closer

import (
	"github.com/C0nstantin/pkg/log"
	"os"
	"os/signal"
	"sync"
)

// Closer ...
type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
}

// New creates a new Closer instance. If one or more os.Signal values are provided as arguments,
// the Closer will automatically call CloseAll when any of the specified signals is received from the OS.
// This allows for graceful shutdown of resources in response to termination signals like SIGINT or SIGTERM.
//
// Parameters:
//
//	sig ...os.Signal - A variadic parameter that allows specifying any number of os.Signal values.
//	                   If provided, the Closer will listen for these signals and initiate CloseAll upon receiving any of them.
//
// Returns:
//
//	*Closer - A pointer to the newly created Closer instance.
func New(sig ...os.Signal) *Closer {
	log.Debugf("new closer")
	c := &Closer{done: make(chan struct{})}
	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			log.Debugf("signal received, closing all")
			signal.Stop(ch)
			c.CloseAll()
			os.Exit(0)
		}()
	}
	return c
}

// Add registers one or more functions to be called by CloseAll. Each function must conform to
// the signature `func() error`. These functions are intended to close or clean up resources
// gracefully when CloseAll is called, typically in response to an application shutdown.
//
// Parameters:
//
//	f ...func() error - A variadic parameter allowing multiple functions to be added. Each function
//	                    should return an error if it fails to close or clean up its resource.
func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// Wait blocks the calling goroutine until the CloseAll method has been called and completed.
// This is useful for ensuring that all resources have been properly closed before the application exits.
// It does not take any parameters and does not return any values.
func (c *Closer) Wait() {
	log.Debugf("wait")
	<-c.done
}

// CloseAll calls all closer functions
func (c *Closer) CloseAll() {
	log.Debugf("close all")
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		// call all Closer funcs async
		errs := make(chan error, len(funcs))
		for _, f := range funcs {
			go func(f func() error) {
				log.Debugf("call closer func")
				errs <- f()
			}(f)
		}

		for i := 0; i < cap(errs); i++ {
			if err := <-errs; err != nil {
				log.Println("error returned from Closer")
			}
		}
	})
	log.Debugf("closed all finished")
}
