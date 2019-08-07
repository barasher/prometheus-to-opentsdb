package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	promHttpC "github.com/prometheus/client_golang/api/prometheus/v1"
	promCommon "github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func buildSimplePair(ts uint64, v float64) promCommon.SamplePair {
	return promCommon.SamplePair{
		Timestamp: promCommon.Time(ts),
		Value:     promCommon.SampleValue(v),
	}
}

func buildMetric(t map[string]string) promCommon.Metric {
	l := len(t)
	m := make(map[promCommon.LabelName]promCommon.LabelValue, l)
	for k, v := range t {
		m[promCommon.LabelName(k)] = promCommon.LabelValue(v)
	}
	return promCommon.Metric(promCommon.LabelSet(m))
}

func buildSampleStream(m promCommon.Metric, v []promCommon.SamplePair) promCommon.SampleStream {
	return promCommon.SampleStream{
		Metric: m,
		Values: v,
	}
}

func getReferenceMatrix() promCommon.Matrix {
	cat1Tags := map[promCommon.LabelName]promCommon.LabelValue{
		promCommon.LabelName("k1"): promCommon.LabelValue("v1"),
		promCommon.LabelName("k2"): promCommon.LabelValue("v2"),
	}
	cat1Values := []promCommon.SamplePair{
		buildSimplePair(1346846400000, 1.2),
		buildSimplePair(1346846401000, 2.4),
	}
	cat1 := promCommon.SampleStream{
		Metric: promCommon.Metric(promCommon.LabelSet(cat1Tags)),
		Values: cat1Values,
	}
	return promCommon.Matrix(
		[]*promCommon.SampleStream{&cat1},
	)
}

func checkOutputMetric(t *testing.T, m OpentsdbMetric, n string, ts uint64, v float32, ta map[string]string) {
	assert.Equal(t, m.Metric, n)
	assert.Equal(t, m.Timestamp, ts)
	assert.Equal(t, m.Value, v)
	assert.Equal(t, len(ta), len(m.Tags))
	for tk, tv := range ta {
		assert.Equal(t, tv, m.Tags[tk])
	}
}

func TestConvertResult(t *testing.T) {
	c := QueryConf{MetricName: "blabla"}
	o, err := Prometheus{}.convertResult(getReferenceMatrix(), c)
	assert.Nil(t, err)
	assert.Len(t, o, 2)
	expTags := map[string]string{"k1": "v1", "k2": "v2"}
	checkOutputMetric(t, o[0], "blabla", 1346846400, 1.2, expTags)
	checkOutputMetric(t, o[1], "blabla", 1346846401, 2.4, expTags)
}

type UnsupportedResult struct{}

func (UnsupportedResult) String() string {
	return "a"
}

func (UnsupportedResult) Type() promCommon.ValueType {
	return promCommon.ValueType(1234)
}

func TestUnsupportedResult(t *testing.T) {
	_, err := Prometheus{}.convertResult(UnsupportedResult{}, QueryConf{})
	assert.NotNil(t, err)
}

func TestDoQueryInputMapping(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := QueryConf{
		Step:  "10m",
		Query: "myQuery",
		Start: start,
		End:   end,
	}
	api := NewPromApiMock()
	check := func(ctx context.Context, query string, r promHttpC.Range) {
		assert.Equal(t, "myQuery", query)
		d, _ := time.ParseDuration("10m")
		assert.Equal(t, d, r.Step)
		assert.Equal(t, start.Unix(), r.Start.Unix())
		assert.Equal(t, end.Unix(), r.End.Unix())
	}
	api.SetQueryRangeCheckFunc(check)
	Prometheus{api: api}.doQuery(ctx, conf)
}

func TestDoQueryError(t *testing.T) {
	api := NewPromApiMock()
	api.SetQueryRangeOutput(nil, nil, fmt.Errorf("a"))
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := QueryConf{
		Step:  "10m",
		Query: "myQuery",
		Start: start,
		End:   end,
	}
	_, _, err := Prometheus{api: api}.doQuery(ctx, conf)
	assert.NotNil(t, err)
}

func TestDoQueryStepParseError(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := QueryConf{
		Step:  "blabla",
		Query: "myQuery",
		Start: start,
		End:   end,
	}
	api := NewPromApiMock()
	_, _, err := Prometheus{api: api}.doQuery(ctx, conf)
	assert.NotNil(t, err)
}

func TestQueryNominal(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := QueryConf{
		MetricName: "blabla",
		Step:       "10m",
		Query:      "myQuery",
		Start:      start,
		End:        end,
	}
	api := NewPromApiMock()
	m := getReferenceMatrix()
	api.SetQueryRangeOutput(m, nil, nil)

	o, err := Prometheus{api: api}.Query(ctx, conf)
	assert.Nil(t, err)
	assert.Len(t, o, 2)
	expTags := map[string]string{"k1": "v1", "k2": "v2"}
	checkOutputMetric(t, o[0], "blabla", 1346846400, 1.2, expTags)
	checkOutputMetric(t, o[1], "blabla", 1346846401, 2.4, expTags)
}

func TestQueryErrorOnQuerying(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := QueryConf{
		MetricName: "blabla",
		Step:       "10m",
		Query:      "myQuery",
		Start:      start,
		End:        end,
	}
	api := NewPromApiMock()
	api.SetQueryRangeOutput(nil, nil, fmt.Errorf("a"))

	_, err := Prometheus{api: api}.Query(ctx, conf)
	assert.NotNil(t, err)
}

func TestNormalize(t *testing.T) {
	s := "1234567890)=azertyuiop^$qsdfghjklm*	<wxcvbn,;:!"
	r := Prometheus{}.normalize(s)
	assert.Equal(t, "1234567890__azertyuiop__qsdfghjklm___wxcvbn____", r)
}
