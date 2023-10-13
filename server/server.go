package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Quotation struct {
	Value     string `json:"ask"`
	Timestamp string `json:"timestamp"`
}

type QuotationResponse struct {
	USDBRL Quotation
}

func getCurrentQuotation() (*Quotation, error) {
	url := os.Getenv("QUOTING_API_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &Quotation{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &Quotation{}, err
	}
	defer res.Body.Close()
	var currentQuotation QuotationResponse
	err = json.NewDecoder(res.Body).Decode(&currentQuotation)
	if err != nil {
		return &Quotation{}, err
	}
	return &currentQuotation.USDBRL, nil
}

func Perform() {
	currentQuotation, err := getCurrentQuotation()
	if err != nil {
		log.Fatalf("Error while trying to get the current quotation: %s", err)
	}
	log.Println(currentQuotation)

	// TODO store the quotation in the database
	// TODO wrap the request in a timeout context (200ms)
	// TODO wrap the storage in a timeout context (10ms)
	// TODO Add a cache layer to read directly from the DB if is the same day
}
