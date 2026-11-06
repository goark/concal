package concal

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goark/errs"
	"github.com/goark/fetch"
)

// SourceType represents the type of the source (file, URL, or unknown).
type SourceType int

const (
	// Source types for the parser
	TypeUnknown SourceType = iota // Unknown source type
	TypeFile                      // File source type
	TypeURL                       // URL source type
)

// Source represents a generic source for calendar data, which can be a file or a URL.
type Source interface {
	Type() SourceType
	Reader() io.Reader
	Close() error
}

// fileSource represents a file-based source for calendar data.
type fileSource struct {
	f *os.File
}

// Type returns the source type for the file source.
func (fs *fileSource) Type() SourceType {
	return TypeFile
}

// Reader returns an io.Reader for the file source.
func (fs *fileSource) Reader() io.Reader {
	return fs.f
}

// Close closes the file source.
func (fs *fileSource) Close() error {
	return fs.f.Close()
}

// openSource opens the given source string as a Source (file or URL) based on its type.
func (p *Parser) openSource(ctx context.Context, s string) (Source, error) {
	if ctx.Err() != nil {
		return nil, errs.Wrap(ctx.Err(), errs.WithContext("src", s))
	}

	src, typ := classifySource(s)
	switch typ {
	case TypeFile:
		// #nosec G304 -- The path is explicitly provided as the calendar source.
		f, err := os.Open(src)
		if err != nil {
			return nil, errs.Wrap(err, errs.WithContext("src", src))
		}
		return &fileSource{f: f}, nil
	case TypeURL:
		return p.openSourceURL(ctx, src)
	default:
		return nil, errs.Wrap(ErrInvalidSource, errs.WithContext("src", s))
	}
}

// urlSource represents a URL-based source for calendar data.
type urlSource struct {
	resp fetch.Response
}

// Type returns the source type for the URL source.
func (us *urlSource) Type() SourceType {
	return TypeURL
}

// Reader returns an io.Reader for the URL source.
func (us *urlSource) Reader() io.Reader {
	return us.resp.Body()
}

// Close closes the URL source.
func (us *urlSource) Close() error {
	return us.resp.Close()
}

// openSourceURL opens the given URL source string as a Source. It uses the context for the HTTP request.
func (p *Parser) openSourceURL(ctx context.Context, src string) (Source, error) {
	if ctx.Err() != nil {
		return nil, errs.Wrap(ctx.Err(), errs.WithContext("src", src))
	}

	u, err := fetch.URL(src)
	if err != nil {
		return nil, errs.Wrap(err, errs.WithContext("url", src))
	}
	resp, err := fetch.New(httpClientOpt(p.forceHTTP11)).GetWithContext(ctx, u)
	if err != nil {
		return nil, errs.Wrap(err, errs.WithContext("url", src))
	}
	return &urlSource{resp: resp}, nil
}

// classifySource classifies the given source string as a file, URL, or unknown type.
func classifySource(s string) (string, SourceType) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", TypeUnknown
	}

	// Web URL check: treat as URL if both scheme and host are present
	if u, err := url.Parse(s); err == nil && u.Scheme != "" && u.Host != "" {
		return u.String(), TypeURL
	}

	// Absolute file path check (e.g., /foo, C:\foo)
	if filepath.IsAbs(s) {
		return s, TypeFile
	}
	abspath, err := filepath.Abs(s) // Convert to absolute path if possible
	if err == nil {
		return abspath, TypeFile
	}

	// other string
	return "", TypeUnknown
}

// httpClientOpt returns the appropriate fetch.ClientOpts based on the flagHTTP11 value.
func httpClientOpt(flagHTTP11 bool) fetch.ClientOpts {
	if flagHTTP11 {
		return fetch.WithHTTP11Client()
	}
	return fetch.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}) // Default HTTP client with 30s timeout
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
