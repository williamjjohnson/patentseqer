package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type response struct {
	Count int
	PatentFileWrapperDataBag []patentFileWrapperDataBag
	RequestIdentifier string
}

type patentFileWrapperDataBag struct {
	ApplicationNumberText string
	PgpubDocumentMetaData documentMetadata
}

type documentMetadata struct {
	FileCreateDateTime string
	FileLocationURI string
	ProductIdentifier string
}

const (
	baseUrlUspto = "https://api.uspto.gov/api/v1/patent/applications/"
	downloadUrlUspto = "https://api.uspto.gov/api/v1/download/applications/"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	apiKeyUspto := os.Getenv("USPTO_API_KEY")
	if apiKeyUspto == "" {
		log.Fatal("USPTO_API_KEY is not set")
	}

	// Produce a json file of base USPTO endpoint for creating struct
	testApplicationNumber := "18713389"
	testPatentUrl := getUsptoEndpoint(testApplicationNumber, "base")
	testPatentResponse, err := requestUspto(apiKeyUspto, testPatentUrl)
	if err != nil {
		log.Fatalf("error requesting patent: %v", err)
	}
	testPatentJsonOut, err := getFormattedJson(testPatentResponse)

	pathPatent := "./patent_response"
	err = os.WriteFile(pathPatent, testPatentJsonOut, 0644)
	
	if err != nil {
		log.Fatalf("error writing patent to file: %v", err)
	}

	// Produce a json file of documents request for creating struct
	testDocumentUrl := getUsptoEndpoint(testApplicationNumber, "documents")
	testDocumentResponse, err := requestUspto(apiKeyUspto, testDocumentUrl)
	testDocumentJsonOut, err := getFormattedJson(testDocumentResponse)
	if err != nil {
		log.Fatalf("error requesting patent: %v", err)
	}

	pathDocument := "./document_response"
	err = os.WriteFile(pathDocument, testDocumentJsonOut, 0644)
	
	if err != nil {
		log.Fatalf("error writing document to file: %v", err)
	}

	// Produce a json file of associated documents request for creating struct
	testAssociatedUrl := getUsptoEndpoint(testApplicationNumber, "associated")
	testAssociatedResponse, err := requestUspto(apiKeyUspto, testAssociatedUrl)
	testAssociatedJsonOut, err := getFormattedJson(testAssociatedResponse)
	if err != nil {
		log.Fatalf("error requesting patent: %v", err)
	}

	pathAssociated := "./associated_response"
	err = os.WriteFile(pathAssociated, testAssociatedJsonOut, 0644)
	
	if err != nil {
		log.Fatalf("error writing associated documents to file: %v", err)
	}

	var initialResponseAssociated response
	
	err = json.Unmarshal(testAssociatedResponse, &initialResponseAssociated)

	if err != nil {
		log.Fatalf("error unmarshaling associated documents initial request to struct: %v", err)
	}

	// Submit second request using the file location URI returned by the first request
	if len(initialResponseAssociated.PatentFileWrapperDataBag) != 1 {
		log.Fatalf("length of response for PatentFileWrapperDataBag has length other than 1")
	}
	testAssociatedUrl = initialResponseAssociated.PatentFileWrapperDataBag[0].PgpubDocumentMetaData.FileLocationURI
	fmt.Printf("Retrieved Url is %v", testAssociatedUrl)
	testAssociatedResponse, err = requestUspto(apiKeyUspto, testAssociatedUrl)
	err = os.WriteFile(pathAssociated, testAssociatedResponse, 0644)

	if err != nil {
		log.Fatalf("error with second request for associated documents: %v", err)
	}

	// Produce a file demonstrating the sequence endpoint response
	testSequenceUrl := getDownloadUsptoEndpoint(testApplicationNumber, "LWKZH8YHREENX12.pdf")
	testSequenceFileResponse, err := requestUspto(apiKeyUspto, testSequenceUrl)

	if err != nil {
		log.Fatalf("error requesting sequence doc: %v", err)
	}

	pathSequenceInitial := "./sequence_response_initial.pdf"
	err = os.WriteFile(pathSequenceInitial, testSequenceFileResponse, 0644)
	
	if err != nil {
		log.Fatalf("error writing document to file: %v", err)
	}
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

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Received status code of: %v\n", resp.StatusCode)
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

