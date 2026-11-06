//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/goark/concal"
)

func main() {
	p := concal.NewParser(
		concal.WithStartTime(time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)),
		concal.WithEndTime(time.Date(2026, time.November, 30, 23, 59, 59, 999999999, time.UTC)),
	)
	evts, err := p.ParseOne(context.Background(), "https://calendar.google.com/calendar/ical/2bk907eqjut8imoorgq1qa4olc@group.calendar.google.com/public/basic.ics")
	// evts, err := p.ParseOne(context.Background(), "../testdata/holiday.ics")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		return
	}
	sort.SliceStable(evts, func(i, j int) bool {
		return evts[i].Start.Before(*evts[j].Start)
	})

	for _, e := range evts {
		fmt.Printf("%s on %s\n", e.Summary, e.Start.Format("2006-01-02"))
	}
}
