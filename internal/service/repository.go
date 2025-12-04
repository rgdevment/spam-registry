package service

import (
	"context"

	"github.com/rgdevment/spam-registry/internal/domain"
)

type Repository interface {
	SaveRawReport(ctx context.Context, r *domain.Report) error

	GetRawReports(ctx context.Context, phoneNumber string) ([]*domain.Report, error)

	UpsertScore(ctx context.Context, s *domain.PhoneScore, ttlSeconds int) error

	UpsertCountryThreat(ctx context.Context, s *domain.PhoneScore, ttlSeconds int) error

	DeleteScore(ctx context.Context, phoneNumber string, countryCode string) error

	GetScore(ctx context.Context, phoneNumber string) (*domain.PhoneScore, error)
}
