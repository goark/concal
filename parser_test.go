package concal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testCalendar = `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test-event
DTSTAMP:20261106T000000Z
DTSTART:20261107T100000Z
DTEND:20261107T110000Z
SUMMARY:Test event
END:VEVENT
END:VCALENDAR
`

func TestParseOne(t *testing.T) {
	calendarFile := filepath.Join(t.TempDir(), "calendar.ics")
	if err := os.WriteFile(calendarFile, []byte(testCalendar), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		source string
	}{
		{name: "file", source: calendarFile},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(testCalendar))
	}))
	defer server.Close()
	tests = append(tests, struct {
		name   string
		source string
	}{name: "url", source: server.URL})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(
				WithStartTime(time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)),
				WithEndTime(time.Date(2026, time.November, 30, 0, 0, 0, 0, time.UTC)),
			)
			events, err := parser.ParseOne(context.Background(), tt.source)
			if err != nil {
				t.Fatalf("ParseOne() error = %v", err)
			}
			if len(events) != 1 {
				t.Fatalf("ParseOne() returned %d events, want 1", len(events))
			}
			if events[0].Summary != "Test event" {
				t.Errorf("event summary = %q, want %q", events[0].Summary, "Test event")
			}
		})
	}
}

func TestParseOneCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parser := NewParser()
	if _, err := parser.ParseOne(ctx, "calendar.ics"); err == nil {
		t.Fatal("ParseOne() error = nil, want canceled context error")
	}
	if parser.GetError() == nil {
		t.Fatal("GetError() = nil, want accumulated error")
	}
}

func TestParseOneHTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	parser := NewParser()
	if _, err := parser.ParseOne(context.Background(), server.URL); err == nil {
		t.Fatal("ParseOne() error = nil, want HTTP status error")
	}
	if parser.GetError() == nil {
		t.Fatal("GetError() = nil, want accumulated error")
	}
}

func TestGetPeriod(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		parser    *Parser
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:   "default period",
			parser: NewParser(),
		},
		{
			name:      "end only",
			parser:    NewParser(WithEndTime(end)),
			wantStart: end.Add(-defaultPeriod),
			wantEnd:   end,
		},
		{
			name:      "start only",
			parser:    NewParser(WithStartTime(start)),
			wantStart: start,
			wantEnd:   start.Add(defaultPeriod),
		},
		{
			name:      "reversed range is swapped",
			parser:    NewParser(WithStartTime(end), WithEndTime(start)),
			wantStart: start,
			wantEnd:   end,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := tt.parser.getPeriod()
			if tt.name == "default period" {
				if gotEnd.Sub(gotStart) != defaultPeriod {
					t.Errorf("default period = %v, want %v", gotEnd.Sub(gotStart), defaultPeriod)
				}
				return
			}
			if !gotStart.Equal(tt.wantStart) || !gotEnd.Equal(tt.wantEnd) {
				t.Errorf("getPeriod() = (%v, %v), want (%v, %v)", gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestGetPeriodDoesNotMutateReversedRange(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	parser := NewParser(WithStartTime(end), WithEndTime(start))
	parser.getPeriod()
	if !parser.start.Equal(end) || !parser.end.Equal(start) {
		t.Errorf("period settings mutated to (%v, %v)", parser.start, parser.end)
	}
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
