package service

import (
	"context"

	"github.com/rgdevment/spam-registry/internal/domain"
)

// Repository defines the contract for data persistence.
// The Service layer doesn't know IF this is Scylla, Postgres, or Memory.
type Repository interface {
	// SaveRawReport stores the forensic evidence.
	SaveRawReport(ctx context.Context, r *domain.Report) error

	// GetScore retrieves the current risk state of a phone number.
	GetScore(ctx context.Context, phoneNumber string) (*domain.PhoneScore, error)

	// UpsertScore updates or inserts a new risk score calculation.
	UpsertScore(ctx context.Context, s *domain.PhoneScore) error
}
