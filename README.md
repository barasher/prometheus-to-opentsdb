# prometheus-to-opentsdb
Query Prometheus, dump results to OoenTSDB

Sample query :

curl 'http://localhost:9090/api/v1/query_range?query=prometheus_http_requests_total%7Bcode!%3D%22302%22%7D&start=2019-07-31T17:00:00.000Z&end=2019-07-31T17:03:00.000Z&step=30s' | jq
prometheus_http_requests_total{code!="302"}