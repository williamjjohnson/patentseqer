package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const (
	baseUrlUspto     = "https://api.uspto.gov/api/v1/patent/applications/"
	downloadUrlUspto = "https://api.uspto.gov/api/v1/download/applications/"
)

var (
	errUnsuccessfulStatusCode  = errors.New("Response returned with unsuccessful status code")
	errIncorrectPatentReturned = errors.New("Api returned a patent with application number not matching the request")
	errEmptyResponseBag        = errors.New("Response returned with empty bag")
	errUnexpectedLength        = errors.New("Response returned from initial request with length other than 1: Behavior Undefined")
	errZeroLengthBuffer        = errors.New("Buffer created with length of zero, indicating empty response")
	apiKeyUspto                string
	outDirectory               = ".out"
)

type response struct {
	Count                    int
	PatentFileWrapperDataBag []patentFileWrapperDataBag
	RequestIdentifier        string
	DocumentBag              []document
}

type patentFileWrapperDataBag struct {
	ApplicationNumberText    string
	PgpubDocumentMetaData    pgpubDocumentMetadata
	ApplicationMetaData      applicationMetaData
	CorrespondenceAddressBag []correspondenceAddress
}

type pgpubDocumentMetadata struct {
	FileCreateDateTime string
	FileLocationURI    string
	ProductIdentifier  string
}

type applicationMetaData struct {
	ApplicantBag                     []applicant
	ApplicationStatusCode            int
	ApplicationStatusDescriptionText string
	CpcClassificationBag             []string
	EarliestPublicationDate          string
	FilingDate                       string
	FirstApplicantName               string
	FirstInventorName                string
	InventionTitle                   string
}

type applicant struct {
	ApplicantNameText        string
	CorrespondenceAddressBag []correspondenceAddress
}

type document struct {
	ApplicationNumberText       string
	DirectionCategory           string
	DocumentCodeDescriptionText string
	DocumentIdentifier          string
	DownloadOptionBag           []downloadOption
	OfficialDate                string
}

type correspondenceAddress struct {
	CityName              string
	CountryCode           string
	CountryName           string
	NameLineOneText       string
	PostalAddressCategory string
}

type downloadOption struct {
	DownloadUrl        string
	MimeTypeIdentifier string
	PageTotalQuantity  int
}

func init() {
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("CI") == "" {
			log.Printf("warning: could not load .env file: %v", err)
		}
	}

	apiKeyUspto = os.Getenv("USPTO_API_KEY")
	if apiKeyUspto == "" {
		log.Fatal("USPTO_API_KEY is not set")
	}
}

func main() {

}

func RequestPatent(applicationNumber string) (response, error) {
	var responseStruct response
	resp, err := fetchPatent(applicationNumber)
	if err != nil {
		return responseStruct, err
	}

	responseStruct, err = responseToStruct(resp)
	if err != nil {
		return responseStruct, err
	}

	err = validatePatentResponse(responseStruct, applicationNumber)
	return responseStruct, err
}

func RequestDocuments(applicationNumber string) (response, error) {
	var responseStruct response
	resp, err := fetchDocumentEndpoint(applicationNumber)
	if err != nil {
		return responseStruct, err
	}

	responseStruct, err = responseToStruct(resp)
	if err != nil {
		return responseStruct, err
	}

	err = validateDocumentsResponse(responseStruct, applicationNumber)
	return responseStruct, err
}

func fetchPatent(applicationNumber string) ([]byte, error) {
	url := getUsptoEndpoint(applicationNumber, "base")
	resp, err := requestUspto(apiKeyUspto, url)
	return resp, err
}

func fetchDocumentEndpoint(applicationNumber string) ([]byte, error) {
	url := getUsptoEndpoint(applicationNumber, "documents")
	resp, err := requestUspto(apiKeyUspto, url)
	return resp, err
}

func fetchAssociatedDocumentEndpoint(applicationNumber string) ([]byte, error) {
	var err error

	url := getUsptoEndpoint(applicationNumber, "associated")
	resp, err := requestUspto(apiKeyUspto, url)
	if err != nil {
		return resp, err
	}

	initialResponseData, err := responseToStruct(resp)
	if err != nil {
		return resp, err
	}

	err = validateInitialResponse(initialResponseData, applicationNumber)
	if err != nil {
		return resp, err
	}

	// Submit second request using the file location URI returned by the first request
	url = initialResponseData.PatentFileWrapperDataBag[0].PgpubDocumentMetaData.FileLocationURI

	resp, err = requestUspto(apiKeyUspto, url)
	return resp, err
}

func validatePatentResponse(resp response, applicationNumber string) error {
	if len(resp.PatentFileWrapperDataBag) == 0 {
		return fmt.Errorf("Error while validating patent response: %w", errEmptyResponseBag)
	}
	if resp.PatentFileWrapperDataBag[0].ApplicationNumberText != applicationNumber {
		return fmt.Errorf("Error while validating patent response: %w", errIncorrectPatentReturned)
	}
	return nil
}

func validateDocumentsResponse(resp response, applicationNumber string) error {
	if len(resp.DocumentBag) == 0 {
		return fmt.Errorf("Error while validating documents response: %w", errEmptyResponseBag)
	}
	if resp.DocumentBag[0].ApplicationNumberText != applicationNumber {
		return fmt.Errorf("Error while validating documents response: %w", errIncorrectPatentReturned)
	}
	return nil
}

func validateInitialResponse(resp response, applicationNumber string) error {
	if len(resp.PatentFileWrapperDataBag) == 0 {
		return fmt.Errorf("Error while validating initial response: %w", errEmptyResponseBag)
	}
	if len(resp.PatentFileWrapperDataBag) != 1 {
		return fmt.Errorf("Error while validating initial response: %w", errUnexpectedLength)
	}
	if resp.PatentFileWrapperDataBag[0].ApplicationNumberText != applicationNumber {
		return fmt.Errorf("Error while validating initial response: %w", errIncorrectPatentReturned)
	}
	return nil
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
		err = errUnsuccessfulStatusCode
		return bodyBytes, fmt.Errorf("%w: %d", err, resp.StatusCode)
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

func responseToJson(bodyBytes []byte, w io.Writer) error {
	formattedData, err := getFormattedJson(bodyBytes)
	if err != nil {
		return err
	}

	_, err = w.Write(formattedData)
	return err
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

func responseToFile(bodyBytes []byte, w io.Writer) error {
	var buf bytes.Buffer
	var err error
	l, _ := buf.Write(bodyBytes)
	buf.WriteTo(w)
	if l < 1 {
		err = errZeroLengthBuffer
	}
	return err
}

func responseToStruct(resp []byte) (response, error) {
	var responseStruct response
	err := json.Unmarshal(resp, &responseStruct)
	return responseStruct, err
}
