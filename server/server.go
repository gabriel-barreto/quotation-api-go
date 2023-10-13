package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Quotation struct {
	ID        string `gorm:"primaryKey"`
	Value     float64
	Timestamp int64 `gorm:"index"`
}

type QuotationResponse struct {
	USDBRL struct {
		Value     string `json:"ask"`
		Timestamp string `json"timestamp"`
	}
}

func createQuotation(r QuotationResponse) (*Quotation, error) {
	value, err := strconv.ParseFloat(r.USDBRL.Value, 0)
	if err != nil {
		return &Quotation{}, nil
	}
	timestamp, err := strconv.ParseInt(r.USDBRL.Timestamp, 10, 0)
	if err != nil {
		return &Quotation{}, nil
	}
	return &Quotation{
		ID:        uuid.New().String(),
		Value:     value,
		Timestamp: timestamp,
	}, nil
}

func setupDB() (*gorm.DB, error) {
	dbName := os.Getenv("SQLITE_DATABASE_NAME")
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Quotation{})
	return db, err
}

func saveQuotation(db *gorm.DB, quotation *Quotation) (*Quotation, error) {
	// instantiating context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// check for the same timestamp
	existingQuotation := Quotation{}
	res := db.
		WithContext(ctx).
		Model(existingQuotation).
		Where("timestamp = ?", quotation.Timestamp).
		Find(&existingQuotation)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 1 {
		return &existingQuotation, nil
	}

	// creating new record
	err := db.WithContext(ctx).Model(&quotation).Create(quotation).Error
	return quotation, err
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
	var quotationResponse QuotationResponse
	err = json.NewDecoder(res.Body).Decode(&quotationResponse)
	if err != nil {
		return &Quotation{}, err
	}
	return createQuotation(quotationResponse)
}

func Perform() {
	db, err := setupDB()
	if err != nil {
		log.Fatalf("Error while trying to connect with the database (%s)", err)
	}
	currentQuotation, err := getCurrentQuotation()
	if err != nil {
		log.Fatalf("Error while trying to get the current quotation: %s", err)
	}
	currentQuotation, err = saveQuotation(db, currentQuotation)
	if err != nil {
		log.Fatalf("Error while trying to store the current quotation: %s", err)
	}
	log.Println(currentQuotation)

	// TODO wrap the request in a timeout context (200ms)
	// TODO wrap the storage in a timeout context (10ms)
	// TODO Add a cache layer to read directly from the DB if is the same day
}
