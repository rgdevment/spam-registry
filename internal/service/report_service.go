package service

import (
	"context"

	"github.com/rgdevment/spam-registry/internal/domain"
)

// reportService is the concrete implementation of the Service interface.
// It is unexported (starts with lowercase) to force usage of the Interface.
type reportService struct {
	repo       Repository
	saltSecret string
}

// NewReportService is the constructor.
// It initializes the logic layer with its necessary dependencies.
func NewReportService(repo Repository, salt string) Service {
	return &reportService{
		repo:       repo,
		saltSecret: salt,
	}
}

// IngestReport implements the logic to validate and save a report.
func (s *reportService) IngestReport(ctx context.Context, rawPhone, country, rawReporter, category, comment string) error {
	// TODO: Implement business logic
	// 1. Validate inputs
	// 2. Hash the reporter phone
	// 3. Create Domain Entity
	// 4. Save to Repo
	return nil
}

// CheckRisk implements the logic to read a score.
func (s *reportService) CheckRisk(ctx context.Context, phoneNumber string) (*domain.PhoneScore, error) {
	// TODO: Call repo to get score
	return nil, nil
}
