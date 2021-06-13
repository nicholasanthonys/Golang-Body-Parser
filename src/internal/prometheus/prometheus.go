package CustomPrometheus

import (
	"github.com/gobeam/stringy"
	"github.com/labstack/gommon/log"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

var (
	PromMapCounter = map[string]prometheus.Counter{}
	Prefix         = "single_middleware_"
	dirProjects    = os.Getenv("CONFIGURES_DIRECTORY_NAME") + "/projects"
)

//func init() {
//	//prometheus.Register(prometheus.NewCounter(prometheus.CounterOpts{
//	//	Name:        "te_ERR_SENDING_REQUEST",
//	//	Help:        "Error Sending Request for related project",
//	//	ConstLabels: prometheus.Labels{"version": "1233"},
//	//}))
//	SetUpMetrics()
//}

func SetUpMetrics(registeredRoutes []model.Route) {

	for _, route := range registeredRoutes {
		setCounterMetricToMap(GetPrefixMetricName(route.MetricPrefixName))
	}
}

func SetPrefixListCounterOptName(prefix string, metrics []prometheus.CounterOpts) []prometheus.CounterOpts {
	for index, metricCounterOpt := range metrics {
		metrics[index].Name = prefix + metricCounterOpt.Name
	}

	log.Infof("returned metrics is")
	log.Info(metrics)
	return metrics
}

func setListCounterOptToMap(opts []prometheus.CounterOpts) {
	for _, opt := range opts {
		PromMapCounter[opt.Name] = prometheus.NewCounter(opt)
		log.Infof("opt Name is %s", opt.Name)
		prometheus.MustRegister(PromMapCounter[opt.Name])
	}
}

func setCounterMetricToMap(prefixMetricName string) {

	serialMetrics := GetSerialCounterMetricOpt()
	appendedSerialMetrics := SetPrefixListCounterOptName(prefixMetricName, serialMetrics)

	parallelMetrics := GetParallelCounterMetricOpt()
	appendedParallelMetrics := SetPrefixListCounterOptName(prefixMetricName, parallelMetrics)

	// general
	generalMetrics := GetGeneralCounterMetricOpt()
	appendedGeneralMetrics := SetPrefixListCounterOptName(prefixMetricName, generalMetrics)
	// request
	requestMetrics := GetRequestCounterMetricOpt()
	appendedRequestMetrics := SetPrefixListCounterOptName(prefixMetricName, requestMetrics)
	// response
	responseMetrics := GetResponseCounterMetricOpt()
	appendedResponseMetrics := SetPrefixListCounterOptName(prefixMetricName, responseMetrics)

	setListCounterOptToMap(appendedSerialMetrics)
	setListCounterOptToMap(appendedParallelMetrics)
	setListCounterOptToMap(appendedGeneralMetrics)
	setListCounterOptToMap(appendedRequestMetrics)
	setListCounterOptToMap(appendedResponseMetrics)
}

func GetPrefixMetricName(metricPrefixName string) string {
	return stringy.New(Prefix+metricPrefixName).SnakeCase("/", "").Get() + "_"

}

func CustomPrometheusHandler() http.Handler {
	return promhttp.Handler()
}
