package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetQueryConf(t *testing.T) {
	defQueryConf := QueryConf{}
	var tcs = []struct {
		tcID         string
		file         string
		expOk        bool
		expQueryConf QueryConf
	}{
		{"emptyFileName", "", false, defQueryConf},
		{"nonExistingFile", "nonExisting.json", false, defQueryConf},
		{"unparsable", "../testdata/unparsable.json", false, defQueryConf},
		{"noMetricName", "../testdata/confFiles/queryConf_noMetricName.json", false, defQueryConf},
		{"noQuery", "../testdata/confFiles/queryConf_noQuery.json", false, defQueryConf},
		{"noStep", "../testdata/confFiles/queryConf_noStep.json", false, defQueryConf},
		{"nominal", "../testdata/confFiles/queryConf_nominal.json", true,
			QueryConf{MetricName: "metricname", Query: "query", Step: "step"},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c, err := GetQueryConf(tc.file)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.expQueryConf.MetricName, c.MetricName)
				assert.Equal(t, tc.expQueryConf.Query, c.Query)
				assert.Equal(t, tc.expQueryConf.Step, c.Step)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestGetExporterConf(t *testing.T) {
	defExporterConf := ExporterConf{}
	var tcs = []struct {
		tcID            string
		file            string
		expOk           bool
		expExporterConf ExporterConf
	}{
		{"emptyFileName", "", false, defExporterConf},
		{"nonExistingFile", "nonExisting.json", false, defExporterConf},
		{"unparsable", "../testdata/unparsable.json", false, defExporterConf},
		{"noPrometheusUrl", "../testdata/confFiles/exporterConf_noPrometheusUrl.json", false, defExporterConf},
		{"nominal", "../testdata/confFiles/exporterConf_nominal.json", true, ExporterConf{"prometheusurl"}},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c, err := GetExporterConf(tc.file)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.expExporterConf.PrometheusURL, c.PrometheusURL)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
