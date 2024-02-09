# salesforce_exporter
A small and simple exporter for getting metrics from Salesforce
# About
It's still a PoC, don't use it in production!
SOQL queries are hardcoded now.
# Configuration
It can be configured via environment variables or command line flags:
| ENV                | Flag                | Description                         |
| -------------------|---------------------| ------------------------------------|
| SFE_LISTEN_ADDRESS | web.listen-address  | Address to listen on for telemetry  |
| SFE_METRICS_PATH   | web.telemetry-path  | Path under which to expose metrics  |
| SFE_SF_URL         | salesforce.url      | Salesforce login URL                |
| SFE_SF_USER        | salesforce.user     | User for interation with Salesforce |
| SFE_SF_PASSWORD    | salesforce.password | User's password                     |
| SFE_SF_TOKEN       | salesforce.token    | User's token                        |
# Provided metrics
| Metrics name              | Metric type   | Description                                |
| --------------------------|---------------| -------------------------------------------|
| salesforce_cases_opened   | gauge         | How many cases have been opened (per type) |
| salesforce_cases_total    | counter       | Total amount of cases                      |
| salesforce_pendings_total | gauge         | Total pending chat requests                |
| salesforce_up             | gauge         | Was the last Salesforce query successful   |
