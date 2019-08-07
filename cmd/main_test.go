package main

import (
	"testing"

	"github.com/sirupsen/logrus"
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
		{"unparsableStartDate", []string{
			"-q", "../testdata/confFiles/queryConf_nominal.json",
			"-e", "../testdata/confFiles/exporterConf_wrongLoggingLevel.json",
			"-f", "2019-07-31T17:00:00.000Z",
			"-t", "2019-07-31T17:03:00.000Z",
		}},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			assert.Equal(t, retConfFailure, doMain(tc.inParams))
		})
	}
}

func TestSetLoggingLevel(t *testing.T) {
	l := logrus.GetLevel()
	var tcs = []struct {
		tcID   string
		inLvl  string
		outLvl logrus.Level
	}{
		{"debug", "debug", logrus.DebugLevel},
		{"info", "info", logrus.InfoLevel},
		{"warn", "warn", logrus.WarnLevel},
		{"error", "error", logrus.ErrorLevel},
		{"fatal", "fatal", logrus.FatalLevel},
		{"panic", "panic", logrus.PanicLevel},
		{"empty", "", logrus.InfoLevel},
		{"caseSensitivity", "PaNiC", logrus.PanicLevel},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			setLoggingLevel(tc.inLvl)
			assert.Equal(t, tc.outLvl, logrus.GetLevel())
		})
	}
	logrus.SetLevel(l)
}
