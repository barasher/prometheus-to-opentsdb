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

// Prometheus is a Prometheus connector
type Prometheus struct {
	api promHttpC.API
}

// NewPrometheus instanciates a Prometheus connector
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

// Query executes the query
func (p Prometheus) Query(ctx context.Context, c QueryConf) ([]OpentsdbMetric, error) {
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

func (p Prometheus) convertResult(v promCommon.Value, c QueryConf) ([]OpentsdbMetric, error) {
	if v.Type() == promCommon.ValMatrix {
		return p.convertMatrix(v.(promCommon.Matrix), c)
	}
	return []OpentsdbMetric{}, fmt.Errorf("unsupported prometheus result type: %v", v.Type())
}

func (p Prometheus) convertMatrix(m promCommon.Matrix, c QueryConf) ([]OpentsdbMetric, error) {
	i := 0
	for _, curSS := range m {
		i += len(curSS.Values)
	}
	out := make([]OpentsdbMetric, i, i)
	logrus.Debugf("%v measures from Prometheus", i)

	i = 0
	for _, curSS := range m {
		tags := p.convertTags(curSS.Metric, c)
		for _, pt := range curSS.Values {
			outCur := OpentsdbMetric{}
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

func (p Prometheus) keepTag(k string, c QueryConf) bool {
	for _, curRem := range c.RemoveTags {
		if curRem == k {
			return false
		}
	}
	return true
}

func (p Prometheus) convertTags(ss promCommon.Metric, c QueryConf) map[string]string {
	tags := make(map[string]string)
	for curTagKey, curTagVal := range ss {
		sK := string(curTagKey)
		sV := string(curTagVal)
		if !p.keepTag(sK, c) { // remove
			continue
		}
		if newK, found := c.RenameTags[sK]; found { // rename
			tags[p.normalize(newK)] = p.normalize(sV)
		} else { // keep unchanged
			tags[p.normalize(sK)] = p.normalize(sV)
		}
	}
	for cK, cV := range c.AddTags { // add
		tags[p.normalize(cK)] = p.normalize(cV)
	}
	return tags
}

func (p Prometheus) normalize(s string) string {
	b := []byte(s)
	for i, c := range b {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') ||
			c == '-' || c == '_' || c == '.' || c == '/') {
			b[i] = '_'
		}
	}
	return string(b)
}
