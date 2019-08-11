# prometheus-to-opentsdb

## Description

**Prometheus-to-Opentsdb** is a tool that executes queries on [Prometheus](https://prometheus.io/) and stores results to [Opentsdb](http://opentsdb.net/).

It is not a remote storage, the typical use-case is metrology.

Here is an example : let's consider a big Kubernetes cluster with a lots of pod running on it. A lots a metrics will be generated, usually consumed by Prometheus. Prometheus is not really designed to store metrics over time, it is rather used to deal with daily supervision. Only a subset of these metrics (eventually downsampled) are relevent to draw trends : here comes **Prometheus-to-Opentsdb**.

It executes use-defined queries on Prometheus and stores the results on a long term time-series storage system : Opentsdb.

## Usage

### Configuration

**Prometheus-to-Opentsdb**'s configuration is divided into three parts.

The **first part**, the __exporter configuration file__ defines the "where": where are the backends ?

```
{
  "PrometheusURL":"http://127.0.0.1:9090",
  "OpentsdbURL":"http://127.0.0.1:4242",
  "LoggingLevel":"debug"
}
```

- __**PrometheusURL**__ defines the Prometheus URL - required
- __**OpentsdbURL**__ defines the Opentsdb URL - required
- __**LoggingLevel**__ defines the logging level (possible values: debug, info, warn, error, fatal, panic) - default value : info

The **second part**, the __query description file__ defines the "what": what's my query and how do I map the results ?

```
{
    "MetricName":"blabla",
    "Query":"prometheus_http_requests_total{code!=\"302\"}",
    "Step":"30s"
}
```

- __**MetricName**__ defines the metric name for the gathered data - required
- __**Query**__ defines the Prometheus query that has to be executed - required
- __**Step**__  defines the step for the Prometheus query - required

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

## Metrics mapping

Metrics, tag names and tag values are normalized to fit Opentsdb constraints. Any character that is not `[a-z]`, `[A-Z]`, `[0-9]` or `_` is replaced by `_`.

## Changelog

- **v1.0** : first (working) version