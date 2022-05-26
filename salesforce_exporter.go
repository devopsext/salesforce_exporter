package main

import (
	"flag"
	"net/http"
	"os"

	"salesforce_exporter/src/salesforce"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const namespace = "salesforce"

var (
	listenAddress string = ":9141"
	metricsPath   string = "/metrics"
	sfURL         string = "https://login.salesforce.com"
	sfUser        string
	sfPassword    string
	sfToken       string

	// Metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last Salesforce query successful.",
		nil, nil,
	)
	casesOpened = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cases_opened"),
		"How many cases have been opened (per type).",
		[]string{
			"sf_case_type",
			"sf_case_origin",
			"sf_case_issue",
			"sf_case_country",
		},
		nil,
	)
	casesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cases_total"),
		"Total amount of cases.",
		nil,
		nil,
	)
)

type Exporter struct {
	sfURL, sfUser, sfPassword, sfToken string
}

func NewExporter(sfURL, sfUser, sfPassword, sfToken string) *Exporter {
	return &Exporter{
		sfURL:      sfURL,
		sfUser:     sfUser,
		sfPassword: sfPassword,
		sfToken:    sfToken,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- casesOpened
	ch <- casesTotal
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	log.Info("Start scraping...")

	sfClient, err := salesforce.CreateClient(
		e.sfURL,
		e.sfUser,
		e.sfPassword,
		e.sfToken,
	)

	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Error(err)
		return
	}

	openedCases, err := salesforce.QueryOpenedCases(sfClient)

	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Error(err)
		return
	}

	totalCases, err := salesforce.QueryTotalCases(sfClient)

	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Error(err)
		return
	}

	for key, value := range openedCases {
		ch <- prometheus.MustNewConstMetric(
			casesOpened,
			prometheus.GaugeValue,
			value,
			key.CaseType,
			key.CaseOrigin,
			key.CaseIssue,
			key.CaseCountry,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		casesTotal, prometheus.CounterValue, totalCases,
	)

	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)

	log.Info("Scraping has successfully finished")

}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {

	flag.StringVar(
		&listenAddress,
		"web.listen-address",
		LookupEnvOrString("SFE_LISTEN_ADDRESS", listenAddress),
		"Address to listen on for telemetry",
	)
	flag.StringVar(
		&metricsPath,
		"web.telemetry-path",
		LookupEnvOrString("SFE_METRICS_PATH", metricsPath),
		"Path under which to expose metrics",
	)
	flag.StringVar(
		&sfURL,
		"salesforce.url",
		LookupEnvOrString("SFE_SF_URL", sfURL),
		"Salesforce login URL",
	)
	flag.StringVar(
		&sfUser,
		"salesforce.user",
		LookupEnvOrString("SFE_SF_USER", sfUser),
		"User for interation with Salesforce",
	)
	flag.StringVar(
		&sfPassword,
		"salesforce.password",
		LookupEnvOrString("SFE_SF_PASSWORD", sfPassword),
		"User's password",
	)
	flag.StringVar(
		&sfToken,
		"salesforce.token",
		LookupEnvOrString("SFE_SF_TOKEN", sfToken),
		"User's token",
	)

	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})

	exporter := NewExporter(sfURL, sfUser, sfPassword, sfToken)
	prometheus.MustRegister(exporter)
	log.Info("Listening on address " + listenAddress)
	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Salesforce Exporter</title></head>
             <body>
             <h1>Salesforce Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatal("Error starting HTTP server")
	}
}
