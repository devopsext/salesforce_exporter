package exporter

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"salesforce_exporter/pkg/salesforce"
)

const namespace = "salesforce"

var (

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

func NewExporter(sfURL, sfUser, sfPassword, sfToken string) (*Exporter, error) {
	if sfURL == "" || sfUser == "" || sfPassword == "" || sfToken == "" {
		return nil, errors.New("no credentials provided")
	}
	return &Exporter{
		sfURL:      sfURL,
		sfUser:     sfUser,
		sfPassword: sfPassword,
		sfToken:    sfToken,
	}, nil
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
