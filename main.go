package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"errors"

	"github.com/joho/godotenv"
)

const (
	baseUrlUspto = "https://api.uspto.gov/api/v1/patent/applications/"
	downloadUrlUspto = "https://api.uspto.gov/api/v1/download/applications/"
)

var (
	errUnsuccessfulStatusCode = errors.New("Response returned with unsuccessful status code")
	errIncorrectPatentReturned = errors.New("Api returned a patent with application number not matching the request")
	apiKeyUspto string
)

type response struct {
	Count int
	PatentFileWrapperDataBag []patentFileWrapperDataBag
	RequestIdentifier string
	DocumentBag []document
}

type patentFileWrapperDataBag struct {
	ApplicationNumberText string
	PgpubDocumentMetaData pgpubDocumentMetadata
	ApplicationMetaData applicationMetaData
	CorrespondenceAddressBag []correspondenceAddress
}

type pgpubDocumentMetadata struct {
	FileCreateDateTime string
	FileLocationURI string
	ProductIdentifier string
}

type applicationMetaData struct {
	ApplicantBag []applicant
	ApplicationStatusCode int
	ApplicationStatusDescriptionText string
	CpcClassificationBag []string
	EarliestPublicationDate string
	FilingDate string
	FirstApplicantName string
	FirstInventorName string
	InventionTitle string
}

type applicant struct {
	ApplicantNameText string 
	CorrespondenceAddressBag []correspondenceAddress
}

type document struct {
	ApplicationNumberText string
	DirectionCategory string
	DocumentCodeDescriptionText string
	DocumentIdentifier string
	DownloadOptionBag []downloadOption
	OfficialDate string
}

type correspondenceAddress struct {
	CityName string
	CountryCode string
	CountryName string
	NameLineOneText string
	PostalAddressCategory string
}

type downloadOption struct {
	DownloadUrl string
	MimeTypeIdentifier string
	PageTotalQuantity int
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	apiKeyUspto = os.Getenv("USPTO_API_KEY")
	if apiKeyUspto == "" {
		log.Fatal("USPTO_API_KEY is not set")
	}
}

func main() {

}

func RequestPatent(applicationNumber string) (response, error) {
	var data response
	var err error 
	
	url := getUsptoEndpoint(applicationNumber, "base")
	resp, err := requestUspto(apiKeyUspto, url)
	if err != nil {
		return data, err
	}
	data, err = responseToStruct(resp)

	if data.PatentFileWrapperDataBag[0].ApplicationNumberText != applicationNumber {
		err = errIncorrectPatentReturned
	}
	return data, err
}

func RequestDocumentEndpoint(applicationNumber string) (response, error) {
	var data response
	var err error 

	url := getUsptoEndpoint(applicationNumber, "documents")
	resp, err := requestUspto(apiKeyUspto, url)
	if err != nil {
		return data, err
	}
	data, err = responseToStruct(resp)
	if data.DocumentBag[0].ApplicationNumberText != applicationNumber {
		err = errIncorrectPatentReturned
	}
	return data, err
}

func RequestAssociatedDocumentEndpoint(applicationNumber string) ([]byte, error) {
	var initialResponseData response
	var err error 

	url := getUsptoEndpoint(applicationNumber, "associated")
	resp, err := requestUspto(apiKeyUspto, url)
	if err != nil {
		return resp, err
	}
	err = json.Unmarshal(resp, &initialResponseData)

	if err != nil {
		return resp, err
	}

	// Submit second request using the file location URI returned by the first request
	if len(initialResponseData.PatentFileWrapperDataBag) != 1 {
		err = errors.New("Response returned from initial request with length other than 1: Behavior Undefined")
		return resp, err
	}

	url = initialResponseData.PatentFileWrapperDataBag[0].PgpubDocumentMetaData.FileLocationURI

	resp, err = requestUspto(apiKeyUspto, url)
	return resp, err
}

func requestUspto(apiKeyUspto string, endpoint string) ([]byte, error) {
	var bodyBytes []byte

	fmt.Printf("Url: %v\n", endpoint)

	req, err := http.NewRequest("GET",
		endpoint, nil)
	if err != nil {
		return bodyBytes, err
	}

	req.Header.Set("X-API-KEY", apiKeyUspto)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bodyBytes, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Received status code of: %v\n", resp.StatusCode)
		err = errUnsuccessfulStatusCode
		return bodyBytes, err
	}
	
	bodyBytes, err = io.ReadAll(resp.Body)
	return bodyBytes, err
}

func getUsptoEndpoint(applicationNumber, requestType string) string {
	var endpoint string
	switch requestType {
	case "base": 
		endpoint = baseUrlUspto + applicationNumber
	case "documents":
		endpoint = baseUrlUspto + applicationNumber + "/documents"
	case "associated":
		endpoint = baseUrlUspto + applicationNumber + "/associated-documents"
	}
	return endpoint
}

func getDownloadUsptoEndpoint(applicationNumber, documentIdentifier string) string {
	downloadEndpoint := downloadUrlUspto + applicationNumber + "/" + documentIdentifier
	return downloadEndpoint
}

func getFormattedJson(bodyBytes []byte) ([]byte, error) {
	var data []byte
	var mapData map[string]any
	err := json.Unmarshal(bodyBytes, &mapData)

	if err != nil {
		return data, err
	}

	data, err = json.MarshalIndent(mapData, "", "	")
	return data, err
}

func responseToStruct(resp []byte) (response, error) {
	var responseStruct response
	err := json.Unmarshal(resp, &responseStruct)	
	return responseStruct, err
}

