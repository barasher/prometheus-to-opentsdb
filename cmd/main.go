package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/barasher/prometheus-to-opentsdb/internal"
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

func main() {
	os.Exit(doMain(os.Args[1:]))
}

func doMain(args []string) int {
	cmd := flag.NewFlagSet("Exporter", flag.ContinueOnError)
	queryConfParam := cmd.String(queryConfParamKey, "", "Query description file")
	exporterConfParam := cmd.String(exporterConfParamKey, "", "Exporter configuration file")
	/*fromParam := cmd.String(fromParamKey, "", "From / start date")
	toParam := cmd.String(toParamKey, "", "To / end date")*/

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
	if *exporterConfParam == "" {
		logrus.Errorf("no exporter configuration file provided provided (-%v)", exporterConfParamKey)
		return retConfFailure
	}

	_, err = internal.GetQueryConf(*queryConfParam)
	if err != nil {
		logrus.Errorf("error while loading query description file '%v': %v", *queryConfParam, err)
		return retConfFailure
	}
	_, err = internal.GetExporterConf(*exporterConfParam)
	if err != nil {
		logrus.Errorf("error while loading exporter configuration file '%v': %v", *exporterConfParam, err)
		return retConfFailure
	}
	
	return retOk
}
