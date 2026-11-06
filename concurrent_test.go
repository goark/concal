package concal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConcurrentWaitClosesEventsAfterLastEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testCalendar))
	}))
	defer server.Close()

	parser := NewParser(
		WithStartTime(time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)),
		WithEndTime(time.Date(2026, time.November, 30, 0, 0, 0, 0, time.UTC)),
	)
	conc := NewConcurrent(context.Background(), parser)
	t.Cleanup(func() {
		if err := conc.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	for range 2 {
		if err := conc.Submit(server.URL); err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	waitDone := make(chan struct{})
	go func() {
		conc.Wait()
		close(waitDone)
	}()

	eventCount := 0
	timeout := time.After(5 * time.Second)
	for {
		select {
		case _, ok := <-conc.Events():
			if !ok {
				if eventCount != 2 {
					t.Fatalf("received %d events, want 2", eventCount)
				}
				return
			}
			eventCount++
		case <-timeout:
			t.Fatalf("timed out waiting for events to close after receiving %d events", eventCount)
		}
	}
}

func TestConcurrentCloseCancelsRequest(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		close(requestStarted)
		<-r.Context().Done()
		close(requestCanceled)
	}))
	defer server.Close()

	conc := NewConcurrent(context.Background(), NewParser())
	defer func() {
		_ = conc.Close()
	}()
	if err := conc.Submit(server.URL); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	select {
	case <-requestStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP request did not start")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- conc.Close()
	}()
	select {
	case err := <-closeDone:
		if err == nil {
			t.Fatal("Close() error = nil, want cancellation error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close() did not return after canceling the request")
	}

	select {
	case <-requestCanceled:
	case <-time.After(5 * time.Second):
		t.Fatal("server request context was not canceled")
	}
}

func TestConcurrentSubmitAfterWait(t *testing.T) {
	conc := NewConcurrent(context.Background(), NewParser())
	conc.Wait()

	if err := conc.Submit("calendar.ics"); err == nil {
		t.Fatal("Submit() error = nil after Wait(), want closed parser error")
	}
	_ = conc.Close()
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
