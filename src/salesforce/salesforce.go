package salesforce

import (
	"log"
	"strings"
	"time"

	"github.com/simpleforce/simpleforce"
)

const minBefore = 5
const openedCasesQuery = "SELECT Case_Issue_Primary__c, Case_Issue_Secondary__c, Type, RecordType.Name FROM Case WHERE CreatedDate > "
const totalCasesQuery = "SELECT COUNT() FROM Case"

type Case struct {
	CaseType   string
	CaseOrigin string
}

var CaseType string
var cases []Case

// Connect to SF
func CreateClient(URL, User, Password, Token string) *simpleforce.Client {

	client := simpleforce.NewClient(URL, simpleforce.DefaultClientID, simpleforce.DefaultAPIVersion)
	if client == nil {
		log.Fatal("Can't connect to SF API")
	}

	err := client.LoginPassword(User, Password, Token)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

// Query only opened cases for last minBefore (5 min by default). Sum by type and origin
func QueryOpenedCases(client *simpleforce.Client) map[Case]float64 {
	now := time.Now().UTC()
	timeBefore := now.Add(time.Duration(-minBefore) * time.Minute).Format(time.RFC3339)

	result, err := client.Query(openedCasesQuery + timeBefore)
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range result.Records {
		if record["Type"] == nil {
			CaseType = "none"
		} else {
			CaseType = record["Type"].(string)
			CaseType = strings.ReplaceAll(CaseType, " ", "_")
		}

		for key, value := range record {
			if key == "RecordType" {
				caseOrigin := value.(map[string]interface{})["Name"]
				caseOrigin = strings.ReplaceAll(caseOrigin.(string), " ", "_")
				cases = append(cases, Case{strings.ToLower(CaseType), strings.ToLower(caseOrigin.(string))})
			}
		}
	}

	// Count uniq cases by type and origin and store them to the map
	casesMap := make(map[Case]float64)
	for _, value := range cases {
		cs := Case{
			CaseOrigin: value.CaseOrigin,
			CaseType:   value.CaseType,
		}
		casesMap[cs] += 1
	}

	return casesMap
}

func QueryTotalCases(client *simpleforce.Client) float64 {
	result, err := client.Query(totalCasesQuery)
	if err != nil {
		log.Fatal(err)
	}

	return float64(result.TotalSize)
}
