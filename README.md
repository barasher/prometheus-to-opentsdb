# prometheus-to-opentsdb
Query Prometheus, dump results to OoenTSDB

Sample query :

curl 'http://localhost:9090/api/v1/query_range?query=prometheus_http_requests_total%7Bcode!%3D%22302%22%7D&start=2019-07-31T17:00:00.000Z&end=2019-07-31T17:03:00.000Z&step=30s' | jq
prometheus_http_requests_total{code!="302"}

{
  "status": "success",
  "data": {
    "resultType": "matrix",
    "result": [
      {
        "metric": {
          "__name__": "prometheus_http_requests_total",
          "code": "200",
          "handler": "/api/v1/label/:name/values",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "values": [
          [
            1564592490,
            "2"
          ],
          [
            1564592520,
            "2"
          ],
          [
            1564592550,
            "3"
          ],
          [
            1564592580,
            "3"
          ]
        ]
      },
      {
        "metric": {
          "__name__": "prometheus_http_requests_total",
          "code": "200",
          "handler": "/api/v1/query",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "values": [
          [
            1564592490,
            "2"
          ],
          [
            1564592520,
            "5"
          ],
          [
            1564592550,
            "9"
          ],
          [
            1564592580,
            "10"
          ]
        ]
      },
      {
        "metric": {
          "__name__": "prometheus_http_requests_total",
          "code": "200",
          "handler": "/api/v1/query_range",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "values": [
          [
            1564592550,
            "1"
          ],
          [
            1564592580,
            "1"
          ]
        ]
      },
      {
        "metric": {
          "__name__": "prometheus_http_requests_total",
          "code": "200",
          "handler": "/graph",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "values": [
          [
            1564592490,
            "2"
          ],
          [
            1564592520,
            "2"
          ],
          [
            1564592550,
            "2"
          ],
          [
            1564592580,
            "2"
          ]
        ]
      },
      {
        "metric": {
          "__name__": "prometheus_http_requests_total",
          "code": "200",
          "handler": "/metrics",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "values": [
          [
            1564592490,
            "2"
          ],
          [
            1564592520,
            "4"
          ],
          [
            1564592550,
            "6"
          ],
          [
            1564592580,
            "8"
          ]
        ]
      },
      {
        "metric": {
          "__name__": "prometheus_http_requests_total",
          "code": "200",
          "handler": "/static/*filepath",
          "instance": "localhost:9090",
          "job": "prometheus"
        },
        "values": [
          [
            1564592490,
            "25"
          ],
          [
            1564592520,
            "25"
          ],
          [
            1564592550,
            "25"
          ],
          [
            1564592580,
            "25"
          ]
        ]
      }
    ]
  }
}

opentsdb body :
[
    {
        "metric": "sys.cpu.nice",
        "timestamp": 1346846400,
        "value": 18,
        "tags": {
           "host": "web01",
           "dc": "lga"
        }
    },
    {
        "metric": "sys.cpu.nice",
        "timestamp": 1346846400,
        "value": 9,
        "tags": {
           "host": "web02",
           "dc": "lga"
        }
    }
]