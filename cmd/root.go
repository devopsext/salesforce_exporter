package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"salesforce_exporter/pkg/exporter"
)

var (
	listenAddress = ":9141"
	metricsPath   = "/metrics"
	sfURL         = "https://login.salesforce.com"
	sfUser        string
	sfPassword    string
	sfToken       string
)

var rootCmd = &cobra.Command{
	Use:   "salesforce_exporter",
	Short: "A small and simple exporter for getting metrics from Salesforce",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.JSONFormatter{})

		exporter, err := exporter.NewExporter(sfURL, sfUser, sfPassword, sfToken)
		if err != nil {
			log.Fatal(err)
		}
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
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func init() {
	flags := rootCmd.PersistentFlags()

	flags.StringVar(
		&listenAddress,
		"web.listen-address",
		getEnv("SFE_LISTEN_ADDRESS", listenAddress),
		"Address to listen on for telemetry",
	)
	flags.StringVar(
		&metricsPath,
		"web.telemetry-path",
		getEnv("SFE_METRICS_PATH", metricsPath),
		"Path under which to expose metrics",
	)
	flags.StringVar(
		&sfURL,
		"salesforce.url",
		getEnv("SFE_SF_URL", sfURL),
		"Salesforce login URL",
	)
	if sfURL == "" {
		log.Fatal("Salesforce URL is required")
	}

	flags.StringVar(
		&sfUser,
		"salesforce.user",
		getEnv("SFE_SF_USER", sfUser),
		"User for integration with Salesforce",
	)
	if sfUser == "" {
		log.Fatal("Salesforce user is required")
	}

	flags.StringVar(
		&sfPassword,
		"salesforce.password",
		getEnv("SFE_SF_PASSWORD", sfPassword),
		"User's password",
	)
	if sfPassword == "" {
		rootCmd.Usage()
		log.Fatal("Salesforce password is required")
	}

	flags.StringVar(
		&sfToken,
		"salesforce.token",
		getEnv("SFE_SF_TOKEN", sfToken),
		"User's token",
	)
	if sfToken == "" {
		log.Fatal("Salesforce token is required")
	}
}
