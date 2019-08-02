package internal

type Prometheus struct{}

func NewPrometheus(c ExporterConf) (*Prometheus, error) {
	return &Prometheus{}, nil
}
