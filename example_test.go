package concal_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/goark/concal"
)

func ExampleParser_ParseOne() {
	p := concal.NewParser(
		concal.WithStartTime(time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)),
		concal.WithEndTime(time.Date(2026, time.November, 30, 23, 59, 59, 999999999, time.UTC)),
	)
	evts, err := p.ParseOne(context.Background(), "testdata/holiday.ics")
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	sort.SliceStable(evts, func(i, j int) bool {
		return evts[i].Start.Before(*evts[j].Start)
	})

	for _, e := range evts {
		fmt.Printf("%s on %s\n", e.Summary, e.Start.Format("2006-01-02"))
	}
	//Output:
	//文化の日 on 2026-11-03
	//勤労感謝の日 on 2026-11-23

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
