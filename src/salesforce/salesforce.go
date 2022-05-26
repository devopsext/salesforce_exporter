package salesforce

import (
	"errors"
	"strings"
	"time"

	"github.com/simpleforce/simpleforce"
)

const minBefore = 5

const openedCasesQuery = "SELECT Case_Issue_Primary__c, Account_Country__c, Type, RecordType.Name FROM Case WHERE IsClosed = false AND CreatedDate >"
const totalCasesQuery = "SELECT COUNT() FROM Case WHERE CreatedDate = TODAY"

type Case struct {
	CaseType    string
	CaseOrigin  string
	CaseIssue   string
	CaseCountry string
}

var CaseType string
var CaseIssue string
var CaseCountry string

// Connect to SF
func CreateClient(URL, User, Password, Token string) (*simpleforce.Client, error) {

	client := simpleforce.NewClient(
		URL,
		simpleforce.DefaultClientID,
		simpleforce.DefaultAPIVersion,
	)
	if client == nil {
		return client, errors.New("can't connect to sf api")
	}

	err := client.LoginPassword(User, Password, Token)
	if err != nil {
		return client, err
	}

	return client, nil
}

// Query only opened cases for last minBefore (5 min by default). Sum by type and origin
func QueryOpenedCases(client *simpleforce.Client) (map[Case]float64, error) {
	var cases []Case

	now := time.Now().UTC()
	timeBefore := now.Add(time.Duration(-minBefore) * time.Minute).Format(time.RFC3339)

	result, err := client.Query(openedCasesQuery + timeBefore)

	if err != nil {
		return make(map[Case]float64), err
	}

	for _, record := range result.Records {
		if record["Type"] == nil {
			CaseType = ""
		} else {
			CaseType = record["Type"].(string)
			CaseType = strings.ReplaceAll(CaseType, " ", "_")
		}

		if record["Case_Issue_Primary__c"] == nil {
			CaseIssue = ""
		} else {
			CaseIssue = record["Case_Issue_Primary__c"].(string)
			CaseIssue = strings.ReplaceAll(CaseIssue, " ", "_")
			_, CaseIssue, _ = strings.Cut(CaseIssue, "-_")
		}

		if record["Account_Country__c"] == nil {
			CaseCountry = ""
		} else {
			CaseCountry = record["Account_Country__c"].(string)
			CaseCountry = strings.ReplaceAll(CaseCountry, " ", "_")
		}

		for key, value := range record {
			if key == "RecordType" {
				caseOrigin := value.(map[string]interface{})["Name"]
				caseOrigin = strings.ReplaceAll(caseOrigin.(string), " ", "_")
				cases = append(
					cases,
					Case{
						strings.ToLower(CaseType),
						strings.ToLower(caseOrigin.(string)),
						strings.ToLower(CaseIssue),
						strings.ToLower(CaseCountry),
					},
				)
			}
		}
	}

	// Count uniq cases by type and origin and store them to the map
	casesMap := make(map[Case]float64)
	for _, value := range cases {
		cs := Case{
			CaseOrigin:  value.CaseOrigin,
			CaseType:    value.CaseType,
			CaseIssue:   value.CaseIssue,
			CaseCountry: value.CaseCountry,
		}
		casesMap[cs] += 1
	}

	return casesMap, nil
}

func QueryTotalCases(client *simpleforce.Client) (float64, error) {
	result, err := client.Query(totalCasesQuery)
	if err != nil {
		return 0, err
	}

	return float64(result.TotalSize), nil
}
