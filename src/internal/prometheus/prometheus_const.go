package CustomPrometheus

import prometheus "github.com/prometheus/client_golang/prometheus"

func GetSerialCounterMetricOpt() []prometheus.CounterOpts {
	return []prometheus.CounterOpts{
		{
			Name:        "ERR_UNMARSHAL_SERIAL_JSON",
			Help:        "Error unmarshalling file serial.json",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
	}
}

func GetParallelCounterMetricOpt() []prometheus.CounterOpts {
	return []prometheus.CounterOpts{
		// parallel specific metrics
		{
			Name:        "ERR_UNMARSHAL_PARALLEL_JSON",
			Help:        "Error unmarshalling file parallel.json",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "PARALLEL_ERR_INVALID_REFER",
			Help:        "Invalid Refer next_success or next_failure which only accept serial.json or parallel.json",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "PARALLEL_ERR_COPY_REQUEST",
			Help:        "Error copy request to map",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "PARALLEL_ERR_COPY_RESPONSE",
			Help:        "Error copy response to map",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "PARALLEL_ERR_COPY_CONFIGURE",
			Help:        "Error copy configure to map",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
	}
}

func GetRequestCounterMetricOpt() []prometheus.CounterOpts {
	return []prometheus.CounterOpts{
		{
			Name:        "ERR_SENDING_REQUEST",
			Help:        "Error Sending Request for related project",
			ConstLabels: prometheus.Labels{"version": "1233"},
		},
		{
			Name:        "SUCCESS_SENDING_REQUEST",
			Help:        "Success Sending Request for related project",
			ConstLabels: prometheus.Labels{"version": "1233"},
		},
		{
			Name:        "ERR_PARSE_REQUEST_JSON",
			Help:        "Error parsing json request to map",
			ConstLabels: prometheus.Labels{"version": "1233"},
		},
		{
			Name:        "ERR_PARSE_REQUEST_XML",
			Help:        "Error parse xml request to map",
			ConstLabels: prometheus.Labels{"version": "1233"},
		},
		{
			Name:        "ERR_TRANSFORM_REQUEST_TO_JSON",
			Help:        "Error parse xml request to map",
			ConstLabels: prometheus.Labels{"version": "1233"},
		},
		{
			Name:        "ERR_TRANSFORM_REQUEST_TO_XML",
			Help:        "Error parse xml request to map",
			ConstLabels: prometheus.Labels{"version": "1233"},
		},
		{
			Name:        "UNKNOWN_REQUEST_CONTENT_TYPE",
			Help:        "Request content type is not application/json or application/xml",
			ConstLabels: prometheus.Labels{"version": "1233"},
		},
	}
}

func GetResponseCounterMetricOpt() []prometheus.CounterOpts {
	return []prometheus.CounterOpts{
		{
			Name:        "ERR_READ_RESPONSE_BODY",
			Help:        "Error read response body",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_CONVERT_RESPONSE_BYTE_TO_JSON",
			Help:        "Error convert response body to json",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_CONVERT_RESPONSE_BYTE_TO_XML",
			Help:        "Error convert response body to xml",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_UNKNOWN_RESPONSE_CONTENT_TYPE",
			Help:        "Unknown transform response which only accept ToJson or ToXml",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
	}
}

func GetGeneralCounterMetricOpt() []prometheus.CounterOpts {
	return []prometheus.CounterOpts{
		{
			Name:        "ERR_GET_WRAPPER",
			Help:        "Error get wrapper from map",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_SET_REQUEST_TO_WRAPPER",
			Help:        "Error set request to wrapper",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_CHECK_CONFIGURE_LOGIC",
			Help:        "Error when checking configure logic",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_APPLY_CONFIGURE_LOGIC",
			Help:        "Error when applying configure logic",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_MARSHALLING_CONFIGURE_LOGIC_RULE",
			Help:        "Error when marshalling configure logic rule",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
		{
			Name:        "ERR_MARSHALLING_CONFIGURE_LOGIC_DATA",
			Help:        "Error when marshalling configure logic data",
			ConstLabels: prometheus.Labels{"version": "1234"},
		},
	}
}
