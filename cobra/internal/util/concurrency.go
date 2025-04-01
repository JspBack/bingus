package util

import (
	"context"
	"sync"
)

type ConcurrencyLimiter struct {
	sem    chan struct{}
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	logger *VerboseLogger
}

func NewConcurrencyLimiter(ctx context.Context, maxConcurrency int) *ConcurrencyLimiter {
	ctx, cancel := context.WithCancel(ctx)
	return &ConcurrencyLimiter{
		sem:    make(chan struct{}, maxConcurrency),
		ctx:    ctx,
		cancel: cancel,
		logger: NewVerboseLogger(ctx),
	}
}

func (c *ConcurrencyLimiter) Execute(fn func()) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.sem <- struct{}{}:
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			defer func() { <-c.sem }()
			fn()
		}()
		return nil
	}
}

func (c *ConcurrencyLimiter) Wait() {
	c.wg.Wait()
}

func (c *ConcurrencyLimiter) Cancel() {
	c.cancel()
}

func (c *ConcurrencyLimiter) Close() {
	c.Wait()
	close(c.sem)
}
