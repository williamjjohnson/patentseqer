package main

import (
	"testing"
)

func TestPatentRequest(t *testing.T) {
	testApplicationNumber := "18713389"

	t.Run("Test Patent Request for USPTO", func(t *testing.T) {
		resp, err := RequestPatent(testApplicationNumber)
		assertNoError(t, err)

		got := resp.PatentFileWrapperDataBag[0].ApplicationMetaData.ApplicantBag[0].CorrespondenceAddressBag[0].PostalAddressCategory
		want := "postal"
		assertString(t, got, want)

		got_int := resp.PatentFileWrapperDataBag[0].ApplicationMetaData.ApplicationStatusCode
		want_int := 41
		assertInt(t, got_int, want_int)

		misidentifiedApplicationNumber := "1234123412341234"
		_, err = RequestPatent(misidentifiedApplicationNumber)
		assertError(t, err)

		nonexistentApplicationNumber := "0000000000"
		_, err = RequestPatent(nonexistentApplicationNumber)
		assertError(t, err)
	})

	t.Run("Test document request for USPTO", func(t *testing.T) {
		resp, err := RequestDocuments(testApplicationNumber)
		assertNoError(t, err)

		got := resp.DocumentBag[1].DownloadOptionBag[0].DownloadUrl
		want := "https://api.uspto.gov/api/v1/download/applications/18713389/MR9NTF8D154X188.pdf"
		assertString(t, got, want)
	})

	t.Run("Request associated document endpoint for USPTO", func(t *testing.T) {
		response, err := fetchAssociatedDocumentEndpoint(testApplicationNumber)
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
	t.Helper()

	if err == nil {
		t.Error("Wanted an error and didn't get one")
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Got an error and didn't want one: %v", err)
	}
}

func assertNotNil(t testing.TB, got []byte) {
	t.Helper()

	if got == nil {
		t.Error("Got nil value and didn't expect one")
	}
}
