package service

import (
	"context"

	"github.com/rgdevment/spam-registry/internal/domain"
)

// Service defines the business logic use cases available to the API/Worker.
type Service interface {
	// IngestReport processes an incoming report request.
	IngestReport(ctx context.Context, rawPhone, country, rawReporter, category, comment string) error

	// CheckRisk retrieves the risk profile for a specific number.
	CheckRisk(ctx context.Context, phoneNumber string) (*domain.PhoneScore, error)
}
