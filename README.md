# Prometheus to Opentsdb exporter

[![Build Status](https://travis-ci.org/barasher/prometheus-to-opentsdb.svg?branch=master)](https://travis-ci.org/barasher/prometheus-to-opentsdb)
[![go report card](https://goreportcard.com/badge/github.com/barasher/go-exiftool "go report card")](https://goreportcard.com/report/github.com/barasher/prometheus-to-opentsdb)
[![GoDoc](https://godoc.org/github.com/barasher/prometheus-to-opentsdb?status.svg)](https://godoc.org/github.com/barasher/prometheus-to-opentsdb)
[![codecov](https://codecov.io/gh/barasher/prometheus-to-opentsdb/branch/master/graph/badge.svg)](https://codecov.io/gh/barasher/prometheus-to-opentsdb)

## Description

**Prometheus-to-Opentsdb** is a tool that executes queries on [Prometheus](https://prometheus.io/) and stores results to [Opentsdb](http://opentsdb.net/).

It is not a remote storage, the typical use-case is metrology.

Here is an example : let's consider a big Kubernetes cluster with a lots of pod running on it. A lots a metrics will be generated, usually consumed by Prometheus. Prometheus is not really designed to store metrics over time, it is rather used to deal with daily supervision. Only a subset of these metrics (eventually downsampled) are relevent to draw trends : here comes **Prometheus-to-Opentsdb**.

It executes use-defined queries on Prometheus, binds data tags and stores the results on a long term time-series storage system : Opentsdb.

## Usage

### Configuration

**Prometheus-to-Opentsdb**'s configuration is divided into three parts.

The **first part**, the __exporter configuration file__ defines the "where": where are the backends ?

```
{
  "PrometheusURL" : "http://127.0.0.1:9090",
  "OpentsdbURL" : "http://127.0.0.1:4242",
  "LoggingLevel" : "debug",
  "BulkSize" : 20
}
```

- __**PrometheusURL**__ defines the Prometheus URL - required
- __**OpentsdbURL**__ defines the Opentsdb URL - required
- __**LoggingLevel**__ defines the logging level (possible values: debug, info, warn, error, fatal, panic) - default value : info
- __**BulkSize**__ defines the size of the bulk pushed to Opentsdb - default value : 50
- __**ThreadCount**__ defines how many goroutines will push data to Opentsdb

The **second part**, the __query description file__ defines the "what": what's my query and how do I map the results ?

```
{
    "MetricName" : "myMetric",
    "Query" : "prometheus_http_requests_total{code!=\"302\"}",
    "Step" : "30s",
    "AddTags" : {
      "aTagNameIWantToAdd" : "aTagValueIWantToAdd"
    },
    "RemoveTags" : [ "aTagNameIDoNotWantToKeepFromPrometheus" ],
    "RenameTags" : {
      "aTagNameIWantToRename" : "aNewTagName"
    }
}
```

- __**MetricName**__ defines the metric name for the gathered data - required
- __**Query**__ defines the Prometheus query that has to be executed - required
- __**Step**__  defines the step for the Prometheus query - required
- Tags are automatically mapped from Prometheus to Opentsdb but it can be tuned :
  - __**AddTags**__  defines the tags that have to be added to the metrics
  - __**RemoveTags**__ defines the tag names that have to be removed for the metrics
  - __**RenameTags**__ defines the tag names that have to be renamed

The **third part** defines all the parameters (command line) relative to a specific execution :
- `-f` and `-t` (both required) defines the date range for the execution. The date format (UTC) is the following `YYYY-MM-DDThh:mm:ss.lllZ` where `YYYY` is the year, `MM` the month, `DD` the day, `hh` the hour, `mm` the minutes, `ss` the seconds and `lll` the milliseconds. Sample : `2019-07-31T17:03:00.000Z`.
- `-s` activates the simulation mode : data will be gathered from Prometheus, mapped as it should be for Opentsdb but it will not be sent but only printed. By default, simulation mode is disabled.

But why such a configuration mechanism ? The objective is in fact :
- to define only one time the "where". You'll probably generate more than one metric from Prometheus : this configuration file will be reused.
- to define only one time each metric definition ("what"), it will certainly be executed more than one time so this configuration file will also be reused
- an execution combines an existing "where", an existing "what" and defines the date range.

### Execution

```
Usage of Exporter:
  -e string
    	Exporter configuration file (where ?)
  -q string
    	Query description file (what ?)
  -f string
    	From / start date (when ?)
  -t string
    	To / end date (when ?)
  -s	Simulation mode (don't push to Opentsdb)
```

Sample:
- `./main -q ~/conf/query.json  -e ~/conf/exporter.conf -f 2019-07-23T00:00:00.000Z -t 2019-07-23T23:59:59.999Z`  : effective execution
- `./main -q ~/conf/query.json  -e ~/conf/exporter.conf -f 2019-07-23T00:00:00.000Z -t 2019-07-23T23:59:59.999Z -s` : simulation

Return codes:
- **0**: everything was fine
- **1**: configuration problem
- **2**: execution problem

## Docker

### Get from docker hub

### Build

`docker build -t barasher/prometheus_to_opentsdb:latest .`

### How to use

Inside the container, configuration files are located :
- `/etc/p2o/exporter.json` for the exporter configuration
- `etc/p2o/query.json` for the query configuration
The files can be injected as volume where executing the container.

It also uses 2 environment variables :
- `P2O_FROM` that defines the start date
- `P2O_TO` that defines the end date

Sample :
```
docker run \
  -v /home/barasher/conf/exporter.conf:/etc/p2o/exporter.json \
  -v /home/barasher/conf/query.json:/etc/p2o/query.json \
  --env P2O_FROM='2019-08-11T13:00:00.000Z' \
  --env P2O_TO='2019-08-11T13:32:00.000Z' \
  --rm \
  barasher/prometheus_to_opentsdb:latest
```

The idea is that :
- you can provide to your clients a "base" Docker image that contains the exporter configuration.
- your clients can build their own Docker image containing the query configuration if they want to (or they can just provide the configuration as volume at each execution)
- your clients executes a Docker image (theirs or yours), specifying the date range for the query.
 



## Metrics mapping

Metrics, tag names and tag values are normalized to fit Opentsdb constraints. Any character that is not `[a-z]`, `[A-Z]`, `[0-9]` or `_` is replaced by `_`.

## Changelog

- **v1.0** : first (working) version
- **[v1.1](https://github.com/barasher/prometheus-to-opentsdb/milestone/1?closed=1)** :
  - add "simulation" mode (does not push to Opentsdb but print data on stdout)
  - allow user to add, rename or remove tags