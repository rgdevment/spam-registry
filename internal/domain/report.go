package domain

import (
	"time"

	"github.com/google/uuid"
)

// RiskCategory represents the specific type of threat reported.
// Using a custom type prevents string typos in the business logic.
type RiskCategory string

// RiskLevel helps clients visualize the danger (e.g., Green, Yellow, Red).
type RiskLevel string

const (
	RiskSpam     RiskCategory = "SPAM"
	RiskFraud    RiskCategory = "FRAUD"
	RiskPhishing RiskCategory = "PHISHING"
	RiskDebt     RiskCategory = "DEBT_COLLECTION"
	RiskSales    RiskCategory = "SALES"

	// RiskAutoBlock is a special system event.
	// It indicates an app automatically blocked a call based on our list.
	// Used for the "Swarm Heuristic" (Velocity Check) but has 0 impact on Score alone.
	RiskAutoBlock RiskCategory = "AUTO_BLOCK"
)

const (
	LevelSafe     RiskLevel = "SAFE"     // Score 0-20
	LevelWarning  RiskLevel = "WARNING"  // Score 21-60
	LevelCritical RiskLevel = "CRITICAL" // Score 61-100
)

// Report is the raw evidence input entity.
// This struct maps to the 'raw_reports' table in ScyllaDB.
type Report struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"` // E.164 format
	CountryCode string    `json:"country_code" db:"country_code"` // ISO 3166-1 alpha-2

	// ReporterHash is the HMAC-SHA256 of the reporter's phone number.
	// We NEVER store the raw reporter phone number for privacy reasons.
	ReporterHash string `json:"reporter_hash" db:"reporter_hash"`

	Category  RiskCategory `json:"category" db:"category"`
	Comment   string       `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
}

// PhoneScore represents the current risk state of a number.
// This struct maps to the 'scores' table in ScyllaDB.
type PhoneScore struct {
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Score       float64   `json:"score" db:"score"` // 0.00 to 100.00
	RiskLevel   RiskLevel `json:"risk_level" db:"risk_level"`

	// LastActivity is crucial for the Decay Algorithm.
	// If this date is old, the Score drops significantly when calculated.
	LastActivity time.Time `json:"last_activity" db:"last_activity"`

	// VelocityHitCount tracks how many AUTO_BLOCK events happened recently.
	// This supports the "Swarm Heuristic" to detect active attacks vs old/passive numbers.
	VelocityHitCount int `json:"velocity_hit_count" db:"velocity_hit_count"`

	TotalReports int `json:"total_reports" db:"total_reports"`
}

// NewReport is a factory to create a clean report instance.
// Note: It expects the ReporterHash to be already calculated by the caller (Service layer).
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
