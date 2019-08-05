package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/barasher/prometheus-to-opentsdb/internal"
	promHttpC "github.com/prometheus/client_golang/api/prometheus/v1"
	promCommon "github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestDoMainConfigurationFailure(t *testing.T) {
	var tcs = []struct {
		tcID     string
		inParams []string
	}{
		{"noQueryConfParam", []string{"-e", "../testdata/confFiles/exporterConf_nominal.json"}},
		{"noExporterConfParam", []string{"-q", "../testdata/confFiles/queryConf_nominal.json"}},
		{"queryConfLoadingFailure", []string{
			"-q", "../testdata/confFiles/queryConf_noQuery.json",
			"-e", "../testdata/confFiles/exporterConf_nominal.json",
		}},
		{"exporterConfLoadingFailure", []string{
			"-q", "../testdata/confFiles/queryConf_nominal.json",
			"-e", "../testdata/confFiles/exporterConf_noPrometheusUrl.json",
		}},
		{"helpParam", []string{"-h"}},
		{"unparsableStartDate", []string{
			"-q", "../testdata/confFiles/queryConf_nominal.json",
			"-e", "../testdata/confFiles/exporterConf_nominal.json",
			"-f", "blabla",
			"-t", "2019-07-31T17:03:00.000Z",
		}},
		{"unparsableStartDate", []string{
			"-q", "../testdata/confFiles/queryConf_nominal.json",
			"-e", "../testdata/confFiles/exporterConf_nominal.json",
			"-f", "2019-07-31T17:00:00.000Z",
			"-t", "blabla",
		}},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			assert.Equal(t, retConfFailure, doMain(tc.inParams))
		})
	}
}

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
		buildSimplePair(42, 1.2),
		buildSimplePair(84, 2.4),
	}
	cat1 := promCommon.SampleStream{
		Metric: promCommon.Metric(promCommon.LabelSet(cat1Tags)),
		Values: cat1Values,
	}
	return promCommon.Matrix(
		[]*promCommon.SampleStream{&cat1},
	)
}

func checkOutputMetric(t *testing.T, m internal.OpenTsdbMetric, n string, ts uint64, v float32, ta map[string]string) {
	assert.Equal(t, m.Metric, n)
	assert.Equal(t, m.Timestamp, ts)
	assert.Equal(t, m.Value, v)
	assert.Equal(t, len(ta), len(m.Tags))
	for tk, tv := range ta {
		assert.Equal(t, tv, m.Tags[tk])
	}
}

func TestConvertResult(t *testing.T) {
	c := internal.QueryConf{MetricName: "blabla"}
	o, err := convertResult(getReferenceMatrix(), c)
	assert.Nil(t, err)
	assert.Len(t, o, 2)
	expTags := map[string]string{"k1": "v1", "k2": "v2"}
	checkOutputMetric(t, o[0], "blabla", 42, 1.2, expTags)
	checkOutputMetric(t, o[1], "blabla", 84, 2.4, expTags)
}

type UnsupportedResult struct{}

func (UnsupportedResult) String() string {
	return "a"
}

func (UnsupportedResult) Type() promCommon.ValueType {
	return promCommon.ValueType(1234)
}

func TestUnsupportedResult(t *testing.T) {
	_, err := convertResult(UnsupportedResult{}, internal.QueryConf{})
	assert.NotNil(t, err)
}

func TestQueryInputMapping(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := internal.QueryConf{
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
	query(ctx, api, conf)
}

func TestQueryError(t *testing.T) {
	api := NewPromApiMock()
	api.SetQueryRangeOutput(nil, nil, fmt.Errorf("a"))
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := internal.QueryConf{
		Step:  "10m",
		Query: "myQuery",
		Start: start,
		End:   end,
	}
	_, _, err := query(ctx, api, conf)
	assert.NotNil(t, err)
}

func TestQueryStepParseError(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(1, 1, 1)
	conf := internal.QueryConf{
		Step:  "blabla",
		Query: "myQuery",
		Start: start,
		End:   end,
	}
	api := NewPromApiMock()
	_, _, err := query(ctx, api, conf)
	assert.NotNil(t, err)
}
