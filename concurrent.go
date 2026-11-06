package concal

import (
	"context"
	"sync"

	"github.com/apognu/gocal"
	"github.com/goark/errs"
)

// Conc parses multiple calendar sources concurrently.
type Conc struct {
	ctx       context.Context
	cancel    context.CancelFunc
	parser    *Parser
	errlist   errs.Errors
	parseWG   sync.WaitGroup
	eventChan chan gocal.Event
	submitMu  sync.Mutex
	closed    bool
	waitOnce  sync.Once
}

// NewConcurrent creates a concurrent parser using parser for each source.
func NewConcurrent(ctx context.Context, parser *Parser) *Conc {
	cctx, ccancel := context.WithCancel(ctx)
	c := &Conc{
		ctx:       cctx,
		cancel:    ccancel,
		parser:    parser,
		errlist:   errs.Errors{},
		eventChan: make(chan gocal.Event),
	}
	return c
}

// Submit starts parsing src. Sources cannot be submitted after Wait or Close.
func (c *Conc) Submit(src string) error {
	c.submitMu.Lock()
	defer c.submitMu.Unlock()
	if c.closed {
		err := errs.Wrap(ErrConcurrentClosed, errs.WithContext("src", src))
		c.errlist.Add(err)
		return err
	}
	if err := c.ctx.Err(); err != nil {
		err = errs.Wrap(err, errs.WithContext("src", src))
		c.errlist.Add(err)
		return err
	}
	c.parseWG.Add(1)
	go func() {
		defer c.parseWG.Done()
		evts, err := c.parser.ParseOne(c.ctx, src)
		if err != nil {
			c.errlist.Add(err)
			return
		}
		for _, evt := range evts {
			select {
			case c.eventChan <- evt:
			case <-c.ctx.Done():
				return
			}
		}
	}()
	return nil
}

// Wait prevents further submissions, waits for all events to be sent, and closes Events.
// Events must be consumed concurrently with Wait to avoid blocking event senders.
func (c *Conc) Wait() {
	c.waitOnce.Do(func() {
		c.submitMu.Lock()
		c.closed = true
		c.submitMu.Unlock()
		c.parseWG.Wait()
		close(c.eventChan)
	})
}

// Events returns the channel of parsed events.
func (c *Conc) Events() <-chan gocal.Event {
	return c.eventChan
}

// EventSlice returns a slice of gocal.Event from the Conc object. If the Conc object is nil, it returns nil.
func (c *Conc) EventSlice() []gocal.Event {
	if c == nil {
		return nil
	}
	evts := []gocal.Event{}
	for e := range c.Events() {
		evts = append(evts, e)
	}
	return evts
}

// Close cancels ongoing parsing, waits for workers, and returns accumulated errors.
func (c *Conc) Close() error {
	c.cancel()
	c.Wait()
	return c.errlist.ErrorOrNil()
}

/* MIT License
 *
 * Copyright 2026 Spiegel
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */
