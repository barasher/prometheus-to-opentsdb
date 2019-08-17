package internal

// OpentsdbMetric describes a metric based on Opentsdb specifications
type OpentsdbMetric struct {
	// Metric is the metric name
	Metric string `json:"metric"`
	// Timestamp is the metric timestamp (UTC timestamp)
	Timestamp uint64 `json:"timestamp"`
	// Value is the value of the Metric
	Value float32 `json:"value"`
	// Tags describes the metric tags
	Tags map[string]string `json:"tags"`
}
