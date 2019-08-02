package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDoMainConfigurationFailure(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inParams  []string
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
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			assert.Equal(t, retConfFailure, doMain(tc.inParams))
		})
	}
}