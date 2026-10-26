package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	baseUrlUspto = "https://api.uspto.gov/api/v1/patent/applications/"
	datetimeForm = "2006-01-02T15:04:05.000-0700"
)

type patent struct {
	applicantBag                     []applicant
	applicationConfirmationNuber     int
	applicationStatusCode            int
	applicationStatusDate            time.Time
	applicationStatusDescriptionText string
	cpcClassificationBag             []cpcClassification
	firstApplicantName               string
	firstInventorName                string
	inventionTitle                   string
	filingDate                       []document
}

type applicant struct {
	applicationNameText      string
	correspondenceAddressBag []address
}

type address struct {
	cityName              string
	countryCode           string
	countryName           string
	nameLineOneText       string
	postalAddressCategory string
}

type cpcClassification struct{}

type document struct {
	applicationNumberText       int
	directionCategory           string
	documentCode                string
	documentCodeDescriptionText string
	documentIdentifier          string
	downloadOptionBag           []downloadOption
	officialDate                time.Time
}

type downloadOption struct {
	downloadUrl        string
	mimeTypeIdentifier string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	apiKeyUspto := os.Getenv("USPTO_API_KEY")
	if apiKeyUspto == "" {
		log.Fatal("USPTO_API_KEY is not set")
	}

	//	testApplicationNumber := "18713389"
	//	testPatentResponse, err := requestUspto(testApplicationNumber, apiKeyUspto, "base")
	if err != nil {
		log.Fatal("error requesting patent")
	}

	//	testDocumentResponse, err := requestUspto(testApplicationNumber, apiKeyUspto, "documents")

	parsedTime, err := time.Parse(datetimeForm, "2024-05-28T15:45:51.000-0400")
	if err != nil {
		log.Fatal(err)
	}

	doc1 := document{
		18713389,
		"INTERNAL",
		"CRFE",
		"Sequence Listing in Computer Readable Format, No errors found",
		"75f103be-5daa-45af-ac55-9db31016bc8b",
		[]downloadOption{
			{
				"https://api.uspto.gov/api/v1/download/applications/18713389/75f103be-5daa-45af-ac55-9db31016bc8b/files/18713389_1.1_validation_report.pdf",
				"PDF",
			},
		},
		parsedTime,
	}
	fmt.Println(doc1)
}

func requestUspto(applicationNumber string, apiKeyUspto string, endpoint string) (map[string]any, error) {
	var patentEndpoint string
	var data map[string]any

	if endpoint == "base" {
		patentEndpoint = baseUrlUspto + applicationNumber
	} else {
		patentEndpoint = baseUrlUspto + applicationNumber + "/" + endpoint
	}

	req, err := http.NewRequest("GET",
		patentEndpoint, nil)
	if err != nil {
		return data, err
	}

	req.Header.Set("X-API-KEY", apiKeyUspto)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return data, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	data, err = unmarshalRequest(bodyBytes)

	return data, err
}

func unmarshalRequest(bodyBytes []byte) (map[string]any, error) {
	var data map[string]any
	err := json.Unmarshal(bodyBytes, &data)

	return data, err
}
