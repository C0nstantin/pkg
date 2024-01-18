// Package closer provides a simple way to close all resources in a graceful way
// based on os.Signal

package closer

import (
	"gitlab.wm.local/wm/pkg/log"
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

// New returns new Closer, if []os.Signal is specified Closer will automatically call CloseAll when one of signals is received from OS
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

// Add func to closer
func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// Wait blocks until all closer functions are done
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
