package echarts

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"os"
	"time"
)

type SeriesSet struct {
	series map[string]map[string]interface{} // series_name | timestamp | value
}

func NewSeriesSet() *SeriesSet {
	return &SeriesSet{
		series: make(map[string]map[string]interface{}),
	}
}

func (s *SeriesSet) AddData(name string, timestamp string, y interface{}) {
	ys, ok := s.series[name]
	if ok {
		ys[timestamp] = y
		s.series[name] = ys
		return
	}

	data := make(map[string]interface{})
	data[timestamp] = y
	s.series[name] = data
}

func lineSmoothArea(title string, xName string, yName string, xAxisTime []string, series *SeriesSet) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: xName,
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: yName,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
	)

	line.SetXAxis(xAxisTime)

	for name, serie := range series.series {
		var ld []opts.LineData

		// single series
		for _, x := range xAxisTime {
			yval, ok := serie[x]
			if !ok {
				yval = 0
			}
			ld = append(ld, opts.LineData{
				Name:  name,
				Value: yval,
			})
		}

		line.AddSeries(name, ld)
	}

	line.SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{
			Show: true,
		}),
		charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: 0.2,
		}),
		charts.WithLineChartOpts(opts.LineChart{
			Smooth:       true,
			ConnectNulls: true,
		}),
	)

	return line
}

type LineCharts struct {
	page *components.Page
}

func NewLineCharts() *LineCharts {
	p := components.NewPage()
	return &LineCharts{
		page: p,
	}
}

func (l *LineCharts) AddChart(title string, xName string, yName string, xAxisTime []string, series *SeriesSet) {
	l.page.AddCharts(lineSmoothArea(title, xName, yName, xAxisTime, series))
}

func (l *LineCharts) Render(id string) error {
	const tSLayout = "2006-01-02T15_04_05"
	f, err := os.Create(fmt.Sprintf("metrics-%s-%s.html", id, time.Now().Format(tSLayout)))
	if err != nil {
		panic(err)
	}

	if err := l.page.Render(io.MultiWriter(f)); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}
	return nil
}
