package CustomPrometheus

import (
	"github.com/gobeam/stringy"
	"github.com/labstack/gommon/log"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

var (
	PromMapCollector = map[string]prometheus.Collector{}
	prefix           = "single_middleware_"
	cpuTemp          = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "cpu_temperature_celsius",
		Help:        "Current temperature of the CPU.",
		ConstLabels: prometheus.Labels{"version": "1234"},
	})
	hdFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "hd_errors_total",
		Help:        "Number of hard-disk errors.",
		ConstLabels: prometheus.Labels{"version": "1234"},
	})

	histogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "golang",
			Name:      "my_histogram",
			Help:      "This is my histogram",
		})

	summary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "golang",
			Name:      "my_summary",
			Help:      "This is my summary",
		})
	dirProjects = os.Getenv("CONFIGURES_DIRECTORY_NAME") + "/projects"
)

func init() {
	setUpMetrics()
}

func setUpMetrics() {
	directories, err := util.GetListFolder(dirProjects)
	if err != nil {
		log.Errorf("Cannot read directory projects  %s,  error : %s", dirProjects, err.Error())
		return
	}
	for _, directory := range directories {
		if directory.IsDir() {
			for _, counterOpt := range GetPrometheusCounterMetricOpts() {
				prefixMetricName := stringy.New(prefix+directory.Name()).SnakeCase("?", "").Get()

				// add prefix for metric opt name
				counterOpt.Name = prefixMetricName + "_" + counterOpt.Name
				PromMapCollector[prefixMetricName] = prometheus.NewCounter(counterOpt)
				prometheus.MustRegister(PromMapCollector[prefixMetricName])
			}

		}
	}
}

func CustomPrometheusHandler() http.Handler {
	return promhttp.Handler()
}
