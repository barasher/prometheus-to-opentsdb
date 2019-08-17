package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpentsdbDefaultValues(t *testing.T) {
	c := ExporterConf{}
	o, err := NewOpentsdb(c)
	assert.Nil(t, err)
	assert.Equal(t, defaultBulkSize, o.bulkSize)
	assert.Equal(t, defaultThreadCount, o.threadCount)
	assert.Equal(t, defaultPushTimeout, o.pushTimeout)
}

func TestNewOpentsdbUnparsableTimeout(t *testing.T) {
	c := ExporterConf{
		PushTimeout: "blabla",
	}
	_, err := NewOpentsdb(c)
	assert.NotNil(t, err)
}

func TestDoPush(t *testing.T) {
	var tcs = []struct {
		tcID       string
		inStatus   int
		inResponse string
		expOK      bool
	}{
		{"nominal", http.StatusOK, "{ \"failed\":0, \"success\":2 }", true},
		{"errorParsable", http.StatusBadRequest, "{ \"failed\":1, \"success\":0 }", false},
		{"errorUnparsable", http.StatusBadRequest, "{", false},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.inStatus)
				io.WriteString(w, tc.inResponse)
			}))
			defer ts.Close()

			ctx := context.TODO()
			c := ExporterConf{OpentsdbURL: ts.URL}
			o, err := NewOpentsdb(c)
			assert.Nil(t, err)

			m := []OpentsdbMetric{
				{
					Metric:    "blabla",
					Timestamp: 42,
					Value:     1.3,
					Tags: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}

			err = o.doPush(ctx, m)
			assert.Equal(t, tc.expOK, err == nil)
		})
	}
}

func TestPushOnError(t *testing.T) {
	i := 0
	httpStatuses := []int{http.StatusBadRequest, http.StatusOK, http.StatusBadRequest}
	pushCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushCount++
		if i >= len(httpStatuses) {
			assert.Fail(t, "out of range")
		}
		w.WriteHeader(httpStatuses[i])
		io.WriteString(w, "{ \"failed\":0, \"success\":2 }")
	}))
	defer ts.Close()

	m := []OpentsdbMetric{
		{Metric: "m1", Timestamp: 42, Value: 1.3},
		{Metric: "m2", Timestamp: 43, Value: 1.4},
		{Metric: "m3", Timestamp: 44, Value: 1.5},
	}

	ctx := context.TODO()
	c := ExporterConf{
		OpentsdbURL: ts.URL,
		BulkSize:    uint(1),
	}
	o, err := NewOpentsdb(c)
	assert.Nil(t, err)

	err = o.Push(ctx, m)
	assert.NotNil(t, err)
	assert.Equal(t, 3, pushCount)
}

func TestPushNominal(t *testing.T) {
	m := []OpentsdbMetric{
		{Metric: "m1", Timestamp: 42, Value: 1.3},
		{Metric: "m2", Timestamp: 43, Value: 1.4},
		{Metric: "m3", Timestamp: 44, Value: 1.5},
		{Metric: "m4", Timestamp: 45, Value: 1.6},
	}

	var tcs = []struct {
		tcID       string
		inBulkSize int
		expBulk    [][]string
		expOK      bool
	}{
		{"1", 1, [][]string{{"m1"}, {"m2"}, {"m3"}, {"m4"}}, true},
		{"2", 2, [][]string{{"m1", "m2"}, {"m3", "m4"}}, true},
		{"3", 3, [][]string{{"m1", "m2", "m3"}, {"m4"}}, true},
		{"4", 4, [][]string{{"m1", "m2", "m3", "m4"}}, true},
		{"5", 5, [][]string{{"m1", "m2", "m3", "m4"}}, true},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			pushedMetricIds := [][]string{}
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// unmarshall query
				pushedMetrics := []OpentsdbMetric{}
				assert.Nil(t, json.NewDecoder(r.Body).Decode(&pushedMetrics))
				curMetricIds := []string{}
				for _, cur := range pushedMetrics {
					curMetricIds = append(curMetricIds, cur.Metric)
				}
				pushedMetricIds = append(pushedMetricIds, curMetricIds)
				// build response
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "{ \"failed\":0, \"success\":2 }")
			}))
			defer ts.Close()

			ctx := context.TODO()
			c := ExporterConf{
				OpentsdbURL: ts.URL,
				BulkSize:    uint(tc.inBulkSize),
				ThreadCount: uint(2),
			}
			o, err := NewOpentsdb(c)
			assert.Nil(t, err)

			err = o.Push(ctx, m)
			assert.Nil(t, err)
			assert.ElementsMatch(t, tc.expBulk, pushedMetricIds)
		})
	}
}
