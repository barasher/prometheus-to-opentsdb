package internal

type OpenTsdbMetric struct {
	Metric string  `json:"metric"`
	Timestamp uint64 `json:"timestamp"`
	Value float32 `json:"value"`
	Tags map[string]string `json:"tags"`
}

