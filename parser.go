package concal

import (
	"context"
	"strings"
	"time"

	"github.com/apognu/gocal"
	"github.com/goark/errs"
)

const defaultPeriod = 90 * 24 * time.Hour // default period of 90 days

// Parser is a struct that holds the context, error list, HTTP client settings, and time period for parsing calendar events.
type Parser struct {
	errlist     errs.Errors
	forceHTTP11 bool
	start       time.Time
	end         time.Time
	locations   map[string]*time.Location
}

// parseOpts is a function type for setting options on the Parser.
type parseOpts func(*Parser)

// WithHTTP11Client sets the HTTP client to use HTTP/1.1.
func WithHTTP11Client(forceHTTP11 bool) parseOpts {
	return func(p *Parser) {
		p.forceHTTP11 = forceHTTP11
	}
}

// WithStartTime sets the start time for the parsing period.
func WithStartTime(start time.Time) parseOpts {
	return func(p *Parser) {
		p.start = start
	}
}

// WithEndTime sets the end time for the parsing period.
func WithEndTime(end time.Time) parseOpts {
	return func(p *Parser) {
		p.end = end
	}
}

// WithLocation sets the time location for the parser.
func WithLocation(name string, location *time.Location) parseOpts {
	return func(p *Parser) {
		name = strings.TrimSpace(name)
		if name != "" {
			p.locations[name] = location
		}
	}
}

// NewParser creates a new Parser instance with the given context and options.
func NewParser(opts ...parseOpts) *Parser {
	p := &Parser{
		errlist:     errs.Errors{},
		forceHTTP11: false,
		locations:   map[string]*time.Location{"UTC": time.UTC},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// GetError returns the aggregated error from the parser, if any.
func (p *Parser) GetError() error {
	return p.errlist.ErrorOrNil()
}

func (p *Parser) getPeriod() (time.Time, time.Time) {
	if p.start.IsZero() && p.end.IsZero() {
		start := time.Now()
		return start, start.Add(defaultPeriod) // default period of 90 days after now
	}
	if p.start.IsZero() && !p.end.IsZero() {
		return p.end.Add(-defaultPeriod), p.end // default period of 90 days before end time
	}
	if !p.start.IsZero() && p.end.IsZero() {
		return p.start, p.start.Add(defaultPeriod) // default period of 90 days after start time
	}
	if p.start.After(p.end) {
		return p.end, p.start
	}
	return p.start, p.end
}

// ParseOne parses the calendar events from the given source string within the specified period. It returns the parsed events and any error encountered.
func (p *Parser) ParseOne(ctx context.Context, s string) (events []gocal.Event, retErr error) {
	src, err := p.openSource(ctx, s)
	if err != nil {
		err = errs.Wrap(err, errs.WithContext("src", s))
		p.errlist.Add(err)
		return nil, err
	}
	defer func() {
		if cerr := src.Close(); cerr != nil {
			cerr = errs.Wrap(cerr, errs.WithContext("src", s))
			p.errlist.Add(cerr)
			retErr = errs.Join(retErr, cerr)
			if retErr != nil {
				events = nil
			}
		}
	}()
	gocal.SetTZMapper(func(s string) (*time.Location, error) {
		s = strings.TrimSpace(s)
		if tzid, ok := p.locations[s]; ok {
			return tzid, nil
		}
		return nil, errs.Wrap(ErrInvalidLocation, errs.WithContext("location", s))
	})
	c := gocal.NewParser(src.Reader())
	start, end := p.getPeriod()
	c.Start = &start
	c.End = &end

	if err := c.Parse(); err != nil {
		err = errs.Wrap(err, errs.WithContext("src", s))
		p.errlist.Add(err)
		return nil, err
	}
	return c.Events, nil
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
