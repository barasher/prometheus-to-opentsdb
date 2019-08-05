package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	queryConfDesc          = "query description"
	queryConfMetricNameKey = "MetricName"
	queryConfQueryKey      = "Query"
	queryConfStepKey       = "Step"

	exporterConfDesc             = "exporter configuration"
	exporterConfPrometheusUrlKey = "PrometheusUrl"
)

func loadJson(f string, i interface{}) error {
	r, err := os.Open(f)
	if err != nil {
		return fmt.Errorf("error when opening file '%v': %v", f, err)
	}
	defer r.Close()

	if err := json.NewDecoder(r).Decode(i); err != nil {
		return fmt.Errorf("error when opening file '%v': %v", f, err)
	}

	return nil
}

func checkNotEmptyString(value string, fieldDesc string, structDesc string) error {
	if value == "" {
		return fmt.Errorf("No %v provided in the %v file", fieldDesc, structDesc)
	}
	return nil
}

type QueryConf struct {
	MetricName string
	Query      string
	Step       string
	Start      time.Time
	End        time.Time
}

func GetQueryConf(f string) (QueryConf, error) {
	c := QueryConf{}
	if err := loadJson(f, &c); err != nil {
		return c, err
	}
	if err := checkNotEmptyString(c.MetricName, queryConfMetricNameKey, queryConfDesc); err != nil {
		return c, err
	}
	if err := checkNotEmptyString(c.Query, queryConfQueryKey, queryConfDesc); err != nil {
		return c, err
	}
	if err := checkNotEmptyString(c.Step, queryConfStepKey, queryConfDesc); err != nil {
		return c, err
	}
	return c, nil
}

type ExporterConf struct {
	PrometheusURL string
}

func GetExporterConf(f string) (ExporterConf, error) {
	c := ExporterConf{}
	if err := loadJson(f, &c); err != nil {
		return c, err
	}
	if err := checkNotEmptyString(c.PrometheusURL, exporterConfPrometheusUrlKey, exporterConfDesc); err != nil {
		return c, err
	}
	return c, nil
}
