package CustomPrometheus

import "github.com/prometheus/client_golang/prometheus"

func GetPrometheusCounterMetricOpts()[]prometheus.CounterOpts{
	return []prometheus.CounterOpts{
		prometheus.CounterOpts{
			Name:       "ERR_SENDING_REQUEST",
			Help:        "Error Sending Request for related project",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
	}
}