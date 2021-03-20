package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var (
	ignoreEmpty = flag.Bool("ignore-empty", true, "")
	numericSort = flag.Bool("numeric-sort", true, "")
	delimiter   = flag.String("delimiter", "\t", "")
)

type app struct {
	grid *ui.Grid
	data map[int]*item
}

type item struct {
	count map[string]float64
	input []string
	bar   *widgets.BarChart
	pie   *widgets.PieChart
	log   *widgets.Paragraph
}

func main() {
	flag.Parse()
	if fi, _ := os.Stdin.Stat(); (fi.Mode() & os.ModeCharDevice) != 0 {
		os.Exit(0)
	}

	if err := ui.Init(); err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	stdin := bufio.NewScanner(os.Stdin)
	stdin.Scan()

	a := &app{}
	a.setup(stdin.Text())

	go func() {
		for stdin.Scan() {
			a.update(stdin.Text())
			ui.Render(a.grid)
		}
	}()

	tick := time.Tick(5 * time.Second)
	ev := ui.PollEvents()

	for {
		select {
		case e := <-ev:
			switch e.Type {
			case ui.ResizeEvent:
				a.redraw()
			case ui.KeyboardEvent:
				switch e.ID {
				case "r", "<Space>":
					a.redraw()
				default: // e.g. "q"
					return // quit
				}
			}
		case <-tick:
			a.redraw()
		}
	}
}

func (a *app) setup(head string) {
	a.data = map[int]*item{}
	n := make([]int, len(strings.Split(head, *delimiter)))
	h := 1.0 / float64(len(n))

	var layer []interface{}
	for i, _ := range n {
		v := &item{
			count: map[string]float64{},
			input: make([]string, 100),
			bar:   widgets.NewBarChart(),
			pie:   widgets.NewPieChart(),
			log:   widgets.NewParagraph(),
		}

		v.bar.Title = fmt.Sprintf(" column %d ", i+1)
		v.bar.BarWidth = 10
		v.log.Text = strings.Join(v.input, "\n")

		a.data[i] = v
		layer = append(layer, ui.NewRow(h,
			ui.NewCol(0.7, v.bar),
			ui.NewCol(0.2, v.pie),
			ui.NewCol(0.1, v.log),
		))
	}

	a.grid = ui.NewGrid()
	a.grid.Set(layer...)
	a.redraw()
}

func (a *app) redraw() {
	w, h := ui.TerminalDimensions()
	a.grid.SetRect(0, 0, w, h)
}

func (a *app) update(input string) {
	for i, v := range strings.Split(input, *delimiter) {
		if len(a.data) < i {
			return
		}
		if v == "" && *ignoreEmpty {
			return
		}

		d := a.data[i]
		d.count[v] += 1
		d.input = append([]string{v}, d.input[:100]...)

		var keys []string
		for k := range d.count {
			keys = append(keys, k)
		}

		if *numericSort {
			sort.Slice(keys, func(i, j int) bool {
				ii, e1 := strconv.Atoi(keys[i])
				jj, e2 := strconv.Atoi(keys[j])
				if e1 != nil || e2 != nil {
					return keys[i] < keys[j]
				}
				return ii < jj
			})
		} else {
			sort.Strings(keys)
		}

		var values []float64
		for _, k := range keys {
			values = append(values, d.count[k])
		}

		d.log.Text = strings.Join(d.input, "\n")
		d.bar.Data = values
		d.bar.Labels = keys

		d.pie.Data = values
		d.pie.LabelFormatter = func(idx int, _ float64) string {
			return keys[idx]
		}
	}
}
