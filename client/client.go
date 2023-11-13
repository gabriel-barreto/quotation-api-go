package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gabriel-barreto/go-quoting-api/shared/models"
)

func getQuotation() (*models.Quotation, error) {
	res, err := http.Get("http://localhost:3000/quote")
	currQuotation := models.Quotation{}
	if err != nil {
		return &currQuotation, err
	}
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&currQuotation)
	return &currQuotation, nil
}

func persistToTxtFile(currQuotation *models.Quotation) error {
	fileContent := fmt.Sprintf("Dólar hoje: USD %.2f", currQuotation.Value)
	err := os.WriteFile("./current-quotation.txt", []byte(fileContent), 0644)
	if err != nil {
		return err
	}
	return nil
}

func Perform() {
	currQuotation, err := getQuotation()
	if err != nil {
		log.Fatalf("unable to get the current quotation: %s", err)
	}
	persistToTxtFile(currQuotation)
	if err != nil {
		log.Fatalf("unable to persist the current quotation: %s", err)
	}
	log.Println("Current quotation persisted!")
}
