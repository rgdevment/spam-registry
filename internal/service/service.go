package service

import (
	"context"

	"github.com/rgdevment/spam-registry/internal/domain"
)

type Service interface {
	IngestReport(ctx context.Context, rawPhone, rawReporter, category, comment string) error

	CheckRisk(ctx context.Context, phoneNumber string) (*domain.PhoneScore, error)

	CalculateAndSaveRisk(ctx context.Context, phoneNumber string) error
}
