//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/goark/concal"
)

func main() {
	p := concal.NewParser(
		concal.WithStartTime(time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)),
		concal.WithEndTime(time.Date(2026, time.November, 30, 23, 59, 59, 999999999, time.UTC)),
	)
	c := concal.NewConcurrent(context.Background(), p)
	defer c.Close()

	c.Submit("../testdata/holiday.ics")
	c.Submit("https://calendar.google.com/calendar/ical/2bk907eqjut8imoorgq1qa4olc@group.calendar.google.com/public/basic.ics")
	go c.Wait() // start waiting in a separate goroutine

	for e := range c.Events() {
		fmt.Printf("%s on %s\n", e.Summary, e.Start.Format("2006-01-02"))
	}
}
