package internal

import (
	"context"
	"fmt"
	"time"

	promC "github.com/prometheus/client_golang/api"
	promHttpC "github.com/prometheus/client_golang/api/prometheus/v1"
	promCommon "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type Prometheus struct {
	api promHttpC.API
}

func NewPrometheus(c ExporterConf) (Prometheus, error) {
	p := Prometheus{}
	promConf := promC.Config{Address: c.PrometheusURL}
	promClient, err := promC.NewClient(promConf)
	if err != nil {
		return p, fmt.Errorf("error while initializing Prometheus http client API: %v", err)
	}
	p.api = promHttpC.NewAPI(promClient)
	return p, nil
}

func (p Prometheus) Query(ctx context.Context, c QueryConf) ([]OpenTsdbMetric, error) {
	v, _, err := p.doQuery(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("error while executing query: %v", err)
	}
	return p.convertResult(v, c)
}

func (p Prometheus) doQuery(ctx context.Context, c QueryConf) (promCommon.Value, promC.Warnings, error) {
	var err error
	var step time.Duration
	if step, err = time.ParseDuration(c.Step); err != nil {
		return nil, nil, fmt.Errorf("error while parsing step (%v): %v", c.Step, err)
	}
	ra := promHttpC.Range{
		Start: c.Start,
		End:   c.End,
		Step:  step,
	}

	return p.api.QueryRange(ctx, c.Query, ra)
}

func (p Prometheus) convertResult(v promCommon.Value, c QueryConf) ([]OpenTsdbMetric, error) {
	if v.Type() == promCommon.ValMatrix {
		return p.convertMatrix(v.(promCommon.Matrix), c)
	}
	return []OpenTsdbMetric{}, fmt.Errorf("unsupported prometheus result type: %v", v.Type())
}

func (p Prometheus) convertMatrix(m promCommon.Matrix, c QueryConf) ([]OpenTsdbMetric, error) {
	i := 0
	for _, curCat := range m {
		i += len(curCat.Values)
	}
	out := make([]OpenTsdbMetric, i, i)
	logrus.Debugf("%v measures from Prometheus", i)

	i = 0
	for _, curCat := range m {
		tags := map[string]string{}
		for curTagKey, curTagVal := range curCat.Metric {
			tags[string(curTagKey)] = string(curTagVal)
		}
		for _, pt := range curCat.Values {
			outCur := OpenTsdbMetric{}
			outCur.Timestamp = uint64(pt.Timestamp) / 1000
			outCur.Value = float32(pt.Value)
			outCur.Tags = tags
			outCur.Metric = c.MetricName
			out[i] = outCur
			i++
		}
	}

	return out, nil
}
