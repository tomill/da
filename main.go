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
	delimiter   = flag.String("delimiter", "\t", "")
	ignoreEmpty = flag.Bool("ignore-empty", true, "Do not show blank value count in a bar chart.")
	numericSort = flag.Bool("numeric-sort", true, "Sort the horizontal axis of a bar chart by numeric.")
	redrawEvery = flag.Int64("redraw-every", 5, "Redraws the screen cleanly every specified number of seconds.")
)

type app struct {
	grid  *ui.Grid
	speed *widgets.Plot
	lines float64
	data  map[int]*item
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

	redraw := time.Tick(time.Duration(*redrawEvery) * time.Second)
	traffic := time.Tick(1 * time.Second)
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
		case <-traffic:
			a.traffic()
		case <-redraw:
			a.redraw()
		}
	}
}

func (a *app) setup(head string) {
	a.data = map[int]*item{}
	n := make([]int, len(strings.Split(head, *delimiter)))
	h := 1.0 / float64(len(n)+1)

	a.speed = widgets.NewPlot()
	a.speed.Title = " / sec "
	a.speed.Data = [][]float64{{0, 0}}
	a.speed.LineColors = []ui.Color{ui.ColorGreen}

	var layer []interface{}
	layer = append(layer, ui.NewRow(h,
		ui.NewCol(1.0, a.speed),
	))

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

func (a *app) traffic() {
	a.speed.Data[0] = append(a.speed.Data[0], a.lines)
	a.lines = 0
}

func (a *app) update(input string) {
	a.lines++
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
