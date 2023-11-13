package models

type Quotation struct {
	ID        string `gorm:"primaryKey"`
	Value     float64
	Timestamp int64 `gorm:"index"`
}

type QuotationResponse struct {
	USDBRL struct {
		Value     string `json:"ask"`
		Timestamp string `json:"timestamp"`
	}
}
