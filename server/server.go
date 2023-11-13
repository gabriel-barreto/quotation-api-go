package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gabriel-barreto/go-quoting-api/shared/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createQuotation(r models.QuotationResponse) (*models.Quotation, error) {
	value, err := strconv.ParseFloat(r.USDBRL.Value, 32)
	if err != nil {
		return &models.Quotation{}, nil
	}
	timestamp, err := strconv.ParseInt(r.USDBRL.Timestamp, 10, 0)
	if err != nil {
		return &models.Quotation{}, nil
	}
	return &models.Quotation{
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
	db.AutoMigrate(&models.Quotation{})
	return db, err
}

func saveQuotation(db *gorm.DB, quotation *models.Quotation) (*models.Quotation, error) {
	// instantiating context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// check for the same timestamp
	existingQuotation := models.Quotation{}
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

func getCurrentQuotation() (*models.Quotation, error) {
	url := os.Getenv("QUOTING_API_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &models.Quotation{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &models.Quotation{}, err
	}
	defer res.Body.Close()
	var quotationResponse models.QuotationResponse
	err = json.NewDecoder(res.Body).Decode(&quotationResponse)
	if err != nil {
		return &models.Quotation{}, err
	}
	return createQuotation(quotationResponse)
}

func getQuotation(db *gorm.DB) (*models.Quotation, error) {
	dbQuotation := &models.Quotation{}
	queryResult := db.Model(dbQuotation).Where("timestamp >= strftime('%s', datetime('now', 'localtime', 'start of day'))").Find(&dbQuotation)
	if queryResult.Error != nil {
		return dbQuotation, queryResult.Error
	}
	if queryResult.RowsAffected == 1 {
		return dbQuotation, nil
	}
	return getCurrentQuotation()
}

func perform() (*models.Quotation, error) {
	db, err := setupDB()
	if err != nil {
		return &models.Quotation{}, err
	}
	currentQuotation, err := getQuotation(db)
	if err != nil {
		return &models.Quotation{}, err
	}
	currentQuotation, err = saveQuotation(db, currentQuotation)
	if err != nil {
		return &models.Quotation{}, err
	}
	return currentQuotation, err
}

func getQuotingController(w http.ResponseWriter, r *http.Request) {
	log.Println("Received HTTP request")
	quotation, err := perform()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Println("Finished quotation perform\nError =>", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	log.Println("Finished quotation perform\nCurrent quotation =>", quotation)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quotation)
}

func Start() {
	log.Println("Starting server")
	mux := http.NewServeMux()
	mux.HandleFunc("/quote", getQuotingController)
	http.ListenAndServe(":3000", mux)
}
