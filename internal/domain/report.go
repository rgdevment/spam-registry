package domain

import (
	"time"

	"github.com/google/uuid"
)

type RiskCategory string

type RiskLevel string

const (
	RiskSpam     RiskCategory = "SPAM"
	RiskFraud    RiskCategory = "FRAUD"
	RiskPhishing RiskCategory = "PHISHING"
	RiskDebt     RiskCategory = "DEBT_COLLECTION"
	RiskSales    RiskCategory = "SALES"

	RiskAutoBlock RiskCategory = "AUTO_BLOCK"
)

const (
	LevelSafe     RiskLevel = "SAFE"     // Score 0-20
	LevelWarning  RiskLevel = "WARNING"  // Score 21-60
	LevelCritical RiskLevel = "CRITICAL" // Score 61-100
)

type Report struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"` // E.164 format
	CountryCode string    `json:"country_code" db:"country_code"` // ISO 3166-1 alpha-2

	ReporterHash string `json:"reporter_hash" db:"reporter_hash"`

	Category  RiskCategory `json:"category" db:"category"`
	Comment   string       `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
}

type PhoneScore struct {
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	CountryCode string    `json:"country_code" db:"country_code"`
	Score       float64   `json:"score" db:"score"` // 0.00 to 100.00
	RiskLevel   RiskLevel `json:"risk_level" db:"risk_level"`

	LastActivity time.Time `json:"last_activity" db:"last_activity"`

	VelocityHitCount int `json:"velocity_hit_count" db:"velocity_hit_count"`

	TotalReports int `json:"total_reports" db:"total_reports"`
}

func NewReport(phone, country, reporterHash string, cat RiskCategory, comment string) *Report {
	return &Report{
		ID:           uuid.New(),
		PhoneNumber:  phone,
		CountryCode:  country,
		ReporterHash: reporterHash,
		Category:     cat,
		Comment:      comment,
		CreatedAt:    time.Now().UTC(),
	}
}
