package main

import (
	"testing"
)

func TestPatentRequest(t *testing.T) {
	testApplicationNumber := "18713389"

	t.Run("Test Patent Request for USPTO", func(t *testing.T) {
		response, err := RequestPatent(testApplicationNumber) 
		assertNoError(t, err)
		got := response.PatentFileWrapperDataBag[0].ApplicationMetaData.ApplicantBag[0].CorrespondenceAddressBag[0].PostalAddressCategory
		want := "postal"
		assertString(t, got, want)
		got_int := response.PatentFileWrapperDataBag[0].ApplicationMetaData.ApplicationStatusCode
		want_int := 41
		assertInt(t, got_int, want_int)
		_, err = RequestPatent("1234123412341234")
		assertError(t, err)
	})

	t.Run("Test document request for USPTO", func(t *testing.T) {
		response, err := RequestDocumentEndpoint(testApplicationNumber)
		assertNoError(t, err)
		got := response.DocumentBag[1].DownloadOptionBag[0].DownloadUrl
		want:= "https://api.uspto.gov/api/v1/download/applications/18713389/MR9NTF8D154X188.pdf"
		assertString(t, got, want)
	})

	t.Run("Request associated document endpoint for USPTO", func(t *testing.T) {
		response, err := RequestAssociatedDocumentEndpoint(testApplicationNumber)
		assertNoError(t, err)
		assertNotNil(t, response)
	}) 
}

func assertInt(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}
}

func assertString(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func assertError(t testing.TB, err error) {
	if err == nil {
		t.Error("Wanted an error and didn't get one")
	}
}

func assertNoError(t testing.TB, err error) {
	if err != nil {
		t.Errorf("Got an error and didn't want one: %v", err)
	}
}

func assertNotNil(t testing.TB, got []byte) {
	if got == nil {
		t.Error("Got nil value and didn't expect one")
	}	
}
