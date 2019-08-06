package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/barasher/prometheus-to-opentsdb/internal"
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

	prometheus, err := internal.NewPrometheus(expConf)
	if err != nil {
		logrus.Errorf("%v", err)
		return retExecFailure
	}

	neutral, err := prometheus.Query(ctx, queryConf)
	if err != nil {
		logrus.Errorf("%v", err)
		return retExecFailure
	}

	j, err := json.Marshal(neutral)
	if err != nil {
		logrus.Errorf("Error while marshaling data: %v", err)
		return retExecFailure
	}
	fmt.Printf("%v", string(j))

	return retOk
}
