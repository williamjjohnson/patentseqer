package main

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestPatentRequest(t *testing.T) {
	t.Run("Test Patent Request for USPTO", func(t *testing.T) {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("error loading .env file")
		}

		apiKeyUspto := os.Getenv("USPTO_API_KEY")
		if apiKeyUspto == "" {
			log.Fatal("USPTO_API_KEY is not set")
		}

		testApplicationNumber := "18713389"

		want := 200
		got := requestUspto(testApplicationNumber, apiKeyUspto, "base").StatusCode

		assertInt(t, got, want)
	})

	t.Run("Test document request for USPTO", func(t *testing.T) {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("error loading .env file")
		}

		apiKeyUspto := os.Getenv("USPTO_API_KEY")
		if apiKeyUspto == "" {
			log.Fatal("USPTO_API_KEY is not set")
		}

		testApplicationNumber := "18713389"

		want := 200
		got := requestUspto(testApplicationNumber, apiKeyUspto, "documents").StatusCode

		assertInt(t, got, want)
	})
}

func assertInt(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}
