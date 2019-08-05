package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/barasher/prometheus-to-opentsdb/internal"
	promC "github.com/prometheus/client_golang/api"
	promHttpC "github.com/prometheus/client_golang/api/prometheus/v1"
	promCommon "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

const (
	retOk          int = 0
	retConfFailure int = 1
	retExecFailure int = 2
)

const (
	queryConfParamKey    string = "q"
	exporterConfParamKey string = "e"
	fromParamKey         string = "f"
	toParamKey           string = "t"
)

const dateFormat string = "2006-01-02T15:04:05.000Z"

func main() {
	os.Exit(doMain(os.Args[1:]))
}

func doMain(args []string) int {
	cmd := flag.NewFlagSet("Exporter", flag.ContinueOnError)
	queryConfParam := cmd.String(queryConfParamKey, "", "Query description file")
	exporterConfParam := cmd.String(exporterConfParamKey, "", "Exporter configuration file")
	fromParam := cmd.String(fromParamKey, "", "From / start date")
	toParam := cmd.String(toParamKey, "", "To / end date")

	ctx := context.Background()

	logrus.SetLevel(logrus.DebugLevel) // TODO rendre configurable

	err := cmd.Parse(args)
	if err != nil {
		if err != flag.ErrHelp {
			logrus.Errorf("error while parsing command line arguments: %v", err)
		}
		return retConfFailure
	}

	if *queryConfParam == "" {
		logrus.Errorf("no query description file provided (-%v)", queryConfParamKey)
		return retConfFailure
	}
	queryConf, err := internal.GetQueryConf(*queryConfParam)
	if err != nil {
		logrus.Errorf("error while loading query description file '%v': %v", *queryConfParam, err)
		return retConfFailure
	}

	if *exporterConfParam == "" {
		logrus.Errorf("no exporter configuration file provided provided (-%v)", exporterConfParamKey)
		return retConfFailure
	}
	expConf, err := internal.GetExporterConf(*exporterConfParam)
	if err != nil {
		logrus.Errorf("error while loading exporter configuration file '%v': %v", *exporterConfParam, err)
		return retConfFailure
	}

	if queryConf.Start, err = time.Parse(dateFormat, *fromParam); err != nil {
		logrus.Errorf("error while parsing start date (%v): %v", *fromParam, err)
		return retConfFailure
	}
	if queryConf.End, err = time.Parse(dateFormat, *toParam); err != nil {
		logrus.Errorf("error while parsing end date (%v): %v", *toParam, err)
		return retConfFailure
	}

	api, err := getPrometheusClient(expConf)
	if err != nil {
		logrus.Errorf("%v", err)
		return retExecFailure
	}

	v, _, err := query(ctx, api, queryConf)
	if err != nil {
		logrus.Errorf("%v", err)
		return retExecFailure
	}

	converted, err := convertResult(v, queryConf)
	if err != nil {
		logrus.Errorf("%v", err)
		return retExecFailure
	}
	fmt.Printf("%v", converted)

	return retOk
}

func getPrometheusClient(c internal.ExporterConf) (promHttpC.API, error) {
	promConf := promC.Config{Address: c.PrometheusURL}
	promClient, err := promC.NewClient(promConf)
	if err != nil {
		return nil, fmt.Errorf("error while initializing Prometheus http client API: %v", err)
	}
	return promHttpC.NewAPI(promClient), nil
}

func query(ctx context.Context, api promHttpC.API, c internal.QueryConf) (promCommon.Value, promC.Warnings, error) {
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
	return api.QueryRange(ctx, c.Query, ra)
}

func convertResult(v promCommon.Value, c internal.QueryConf) ([]internal.OpenTsdbMetric, error) {
	if v.Type() == promCommon.ValMatrix {
		return convertMatrix(v.(promCommon.Matrix), c)
	}
	return []internal.OpenTsdbMetric{}, fmt.Errorf("Unsupported prometheus result type: %v", v.Type())
}

func convertMatrix(m promCommon.Matrix, c internal.QueryConf) ([]internal.OpenTsdbMetric, error) {
	i := 0
	for _, curCat := range m {
		i += len(curCat.Values)
	}
	out := make([]internal.OpenTsdbMetric, i, i)
	logrus.Debugf("%v measures from Prometheus", i)

	i = 0
	for _, curCat := range m {
		tags := map[string]string{}
		for curTagKey, curTagVal := range curCat.Metric {
			tags[string(curTagKey)] = string(curTagVal)
		}
		for _, pt := range curCat.Values {
			outCur := internal.OpenTsdbMetric{}
			outCur.Timestamp = uint64(pt.Timestamp)
			outCur.Value = float32(pt.Value)
			outCur.Tags = tags
			outCur.Metric = c.MetricName
			out[i] = outCur
			i++
		}
	}

	return out, nil
}
