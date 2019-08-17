package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
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
	simuParamKey         string = "s"
)

const dateFormat string = "2006-01-02T15:04:05.000Z"

const defaultLoggingLevel string = "info"

var loggingLevels = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
	"fatal": logrus.FatalLevel,
	"panic": logrus.PanicLevel,
}

func setLoggingLevel(s string) error {
	if s == "" {
		logrus.SetLevel(logrus.InfoLevel)
		return nil
	}
	lvl, found := loggingLevels[strings.ToLower(s)]
	if !found {
		return fmt.Errorf("Wrong logging level value (%v)", s)
	}
	logrus.SetLevel(lvl)
	return nil
}

func main() {
	os.Exit(doMain(os.Args[1:]))
}

func doMain(args []string) int {
	cmd := flag.NewFlagSet("Exporter", flag.ContinueOnError)
	queryConfParam := cmd.String(queryConfParamKey, "", "Query description file")
	exporterConfParam := cmd.String(exporterConfParamKey, "", "Exporter configuration file")
	fromParam := cmd.String(fromParamKey, "", "From / start date")
	toParam := cmd.String(toParamKey, "", "To / end date")
	simuParam := cmd.Bool(simuParamKey, false, "Simulation mode (don't push to Opentsdb)")

	ctx := context.Background()

	logrus.SetLevel(logrus.DebugLevel) // TODO rendre configurable

	err := cmd.Parse(args)
	if err != nil {
		if err != flag.ErrHelp {
			logrus.Errorf("error while parsing command line arguments: %v", err)
		}
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
	if setLoggingLevel(expConf.LoggingLevel) != nil {
		logrus.Errorf("%v", err)
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
		logrus.Errorf("error while creating prometheus connector: %v", err)
		return retExecFailure
	}
	neutral, err := prometheus.Query(ctx, queryConf)
	if err != nil {
		logrus.Errorf("%v", err)
		return retExecFailure
	}

	if *simuParam { // simulation mode
		j, err := json.MarshalIndent(neutral, "", "\t")
		if err != nil {
			logrus.Errorf("error while printing results: %v", err)
			return retExecFailure
		}
		fmt.Printf("%v\n", string(j))
		return retOk
	}

	opentsdb, err := internal.NewOpentsdb(expConf)
	if err != nil {
		logrus.Errorf("error while creating opentsdb connector: %v", err)
		return retExecFailure
	}
	if err := opentsdb.Push(ctx, neutral); err != nil {
		logrus.Errorf("%v", err)
		return retExecFailure
	}

	return retOk
}
