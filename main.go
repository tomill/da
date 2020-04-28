package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const Delimiter = "\t"

type app struct {
	grid *ui.Grid
	data map[int]*item
}

type item struct {
	total map[string]float64
	bar   *widgets.BarChart
	pie   *widgets.PieChart
	log   *widgets.Paragraph
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	fmt.Print("loading...")

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

	ev := ui.PollEvents()
	for {
		select {
		case e := <-ev:
			switch e.Type {
			case ui.ResizeEvent:
				a.display()
			case ui.KeyboardEvent:
				return // quit
			}
		}
	}
}

func (a *app) setup(head string) {
	a.data = map[int]*item{}
	n := make([]int, len(strings.Split(head, Delimiter)))
	h := 1.0 / float64(len(n))

	var layer []interface{}
	for i, _ := range n {
		v := &item{
			total: map[string]float64{},
			bar:   widgets.NewBarChart(),
			pie:   widgets.NewPieChart(),
			log:   widgets.NewParagraph(),
		}

		v.bar.Title = fmt.Sprint(i)
		v.bar.BarWidth = 10

		a.data[i] = v
		layer = append(layer, ui.NewRow(h,
			ui.NewCol(0.7, v.bar),
			ui.NewCol(0.2, v.pie),
			ui.NewCol(0.1, v.log),
		))
	}

	a.grid = ui.NewGrid()
	a.grid.Set(layer...)
	a.display()
}

func (a *app) display() {
	w, h := ui.TerminalDimensions()
	a.grid.SetRect(0, 0, w, h)
}

func (a *app) update(input string) {
	for i, v := range strings.Split(input, Delimiter) {
		if len(a.data) < i {
			return
		}

		d := a.data[i]
		d.total[v] += 1

		var keys []string
		for k := range d.total {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var values []float64
		for _, k := range keys {
			values = append(values, d.total[k])
		}

		d.bar.Data = values
		d.bar.Labels = keys

		d.pie.Data = values
		d.pie.LabelFormatter = func(idx int, _ float64) string {
			return keys[idx]
		}

		if len(d.log.Text) > 100 {
			d.log.Text = d.log.Text[:100]
		}
		d.log.Text = v + "\n" + d.log.Text
	}
}
