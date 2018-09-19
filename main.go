package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"

	//"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"sync"
)

var (
	listeningAddress = flag.String("telemetry.address", ":9309", "Address on which to expose metrics.")
	metricsEndpoint  = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metric.")
	showVersion      = flag.Bool("version", false, "Print version information.")
)

const (
	namespace = "puppet_last_run_exporter"
)


type Exporter struct {
	mutex  sync.Mutex

	resourcesChanged          *prometheus.Desc
	resourcesCorrectiveChange *prometheus.Desc
	resourcesFailed	    	  *prometheus.Desc
	resourcesFailedRestart	  *prometheus.Desc
	resourcesOutOfSync        *prometheus.Desc
	resourcesRestarted  	  *prometheus.Desc
	resourcesScheduled		  *prometheus.Desc
	resourcesSkipped		  *prometheus.Desc
	resourcesTotal			  *prometheus.Desc

}


type T struct {
	Version struct {
		Config string `yaml: "version_config"`
		Puppet string `yaml: "version_puppet"`
	}

	Resources struct {
		Changed           float64 `yaml: "resources_changed"`
		Corrective_change float64 `yaml: "resouces_corrective_change"`
		Failed            float64
		Failed_to_restart float64
		Out_of_sync       float64
		Restarted         float64
		Scheduled         float64
		Skipped           float64
		Total             float64
	}

	Time struct {
		Anchor float64
		Archive float64
		Catalog_application float64
		Config_retrieval float64
		Convert_catalog float64
		Exec float64
		Fact_generation float64
		File float64
		Filebucket float64
		Group float64
		Node_retrieval float64
		//package float64 `yaml: "package_resource"`
		Plugin_sync  float64
		Schedule float64
		Service float64
		Total float64
		Transaction_evaluation float64
		User float64
		Yumrepo float64
		Last_run float64
	}
	Changes struct {
		Changes float64
		Total float64
	}
	Events struct {
		Failure float64
		Success float64
		Total float64
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("puppet_last_run_exporter"))
}

/*func NewVersionExporter() *Exporter {
	return &Exporter{
		versionPuppet: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Version"),
			"Puppet versions",
			nil,
			nil,
			),

	}
} */

func NewResourcesExporter() *Exporter {
	return &Exporter{
		resourcesChanged: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "ResourcesChanged"),
			"Number of changed resources",
			nil,
			nil,
		),
		resourcesCorrectiveChange: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "ResourcesCorrectiveChange"),
			"Number of corrective changes",
			nil,
			nil,
		),
		resourcesTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "ResourcesTotal"),
			"Total number of resources",
			nil,
			nil,
		),
	}


}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Printf("Error scraping puppet_last_run_report: %s", err)
	}
	return
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {

	dat, error := ioutil.ReadFile("last_run_summary.yaml")
	if error != nil {
		log.Fatal(error)
	}

	var t T

	err := yaml.Unmarshal(dat, &t)
	if err != nil {
		panic(err)
	}

	ch <- prometheus.MustNewConstMetric(e.resourcesChanged, prometheus.GaugeValue, t.Resources.Changed)
	ch <- prometheus.MustNewConstMetric(e.resourcesCorrectiveChange, prometheus.GaugeValue, t.Resources.Corrective_change)
	ch <- prometheus.MustNewConstMetric(e.resourcesTotal, prometheus.GaugeValue, t.Resources.Total)
	return nil
}


func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.resourcesChanged
	ch <- e.resourcesCorrectiveChange
	ch <- e.resourcesTotal
}


func main(){


	// versionExporter := NewVersionExporter()
	resourceExporter := NewResourcesExporter()
	prometheus.MustRegister(resourceExporter)

	log.Printf("Starting Server: %s", *listeningAddress)
	http.Handle(*metricsEndpoint, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listeningAddress, nil))


}

