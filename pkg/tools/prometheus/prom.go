package prometheus

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/prom2json"
	"io"
	"strconv"
	"time"
)

func ToFamily(in io.Reader) (map[string]*prom2json.Family, error) {
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		return nil, errors.WithMessage(err, "reading text format failed")
	}
	family := make(map[string]*prom2json.Family)
	for _, val := range metricFamilies {
		// add timestamp in metrics
		for _, m := range val.Metric {
			timeMs := time.Now().Unix()*1e3 + int64(time.Now().Nanosecond())/1e6
			m.TimestampMs = &timeMs
		}

		f := prom2json.NewFamily(val)
		family[f.Name] = f
	}
	return family, nil
}

func GetMiB(metric map[string]*prom2json.Family, name string) (float64, error) {
	bytes, ok := metric[name]
	if !ok {
		return 0, nil
	}
	for _, m := range bytes.Metrics {
		promMetric := m.(prom2json.Metric)

		floatVal, err := strconv.ParseFloat(promMetric.Value, 64)
		if err != nil {
			return 0, err
		}
		mib := floatVal / 1024 / 1024
		val, err := strconv.ParseFloat(fmt.Sprintf("%.2f", mib), 64)
		if err != nil {
			return 0, err
		}
		return val, nil
	}
	return 0, nil
}

func SingleGAUGE(metric map[string]*prom2json.Family, name string) (float64, error) {
	gauge, ok := metric[name]
	if !ok {
		return 0, nil
	}

	for _, m := range gauge.Metrics {
		promMetric := m.(prom2json.Metric)
		floatVal, err := strconv.ParseFloat(promMetric.Value, 64)
		if err != nil {
			return 0, err
		}
		return floatVal, nil
	}

	return 0, nil
}

type MultiSeries struct {
	Name  string
	Value float64
}

func NewMultiSeries(name string, value float64) *MultiSeries {
	return &MultiSeries{
		Name:  name,
		Value: value,
	}
}

func GetLoggieMultiGAUGE(metric map[string]*prom2json.Family, name string) ([]*MultiSeries, error) {
	gauge, ok := metric[name]
	if !ok {
		return nil, nil
	}

	var series []*MultiSeries
	for _, m := range gauge.Metrics {
		promMetric := m.(prom2json.Metric)
		seriesName := fmt.Sprintf("%s/%s", promMetric.Labels["pipeline"], promMetric.Labels["source"])
		floatVal, err := strconv.ParseFloat(promMetric.Value, 64)
		if err != nil {
			return nil, err
		}
		series = append(series, NewMultiSeries(seriesName, floatVal))
	}

	return series, nil
}
