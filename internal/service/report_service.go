package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
	"github.com/rgdevment/spam-registry/internal/domain"
)

type reportService struct {
	repo       Repository
	saltSecret string
}

func NewReportService(repo Repository, salt string) Service {
	return &reportService{
		repo:       repo,
		saltSecret: salt,
	}
}

func (s *reportService) IngestReport(ctx context.Context, rawPhone, rawReporter, category, comment string) error {
	num, err := phonenumbers.Parse(rawPhone, "")
	if err != nil {
		return errors.New("invalid phone format: ensure it includes country code (e.g. +569...)")
	}

	if !phonenumbers.IsValidNumber(num) {
		return errors.New("invalid phone number: number does not exist")
	}

	isoRegion := phonenumbers.GetRegionCodeForNumber(num)
	if isoRegion == "" {
		return errors.New("could not detect country from phone number")
	}

	cleanPhone := phonenumbers.Format(num, phonenumbers.E164)

	if rawReporter == "" {
		return errors.New("reporter identity is missing")
	}
	reporterHash := s.generateHash(rawReporter)

	riskCat := domain.RiskCategory(strings.ToUpper(category))

	report := domain.NewReport(
		cleanPhone,
		isoRegion,
		reporterHash,
		riskCat,
		comment,
	)

	return s.repo.SaveRawReport(ctx, report)
}

func (s *reportService) CheckRisk(ctx context.Context, phoneNumber string) (*domain.PhoneScore, error) {
	return s.repo.GetScore(ctx, phoneNumber)
}

func (s *reportService) generateHash(input string) string {
	h := hmac.New(sha256.New, []byte(s.saltSecret))
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *reportService) CalculateAndSaveRisk(ctx context.Context, phoneNumber string) error {
	history, err := s.repo.GetRawReports(ctx, phoneNumber)
	if err != nil {
		return err
	}

	if len(history) == 0 {
		return s.repo.DeleteScore(ctx, phoneNumber, "XX")
	}

	const (
		HalfLifeDays   = 110.0
		OneYearSeconds = 31536000
	)

	weights := map[domain.RiskCategory]float64{
		domain.RiskFraud:     100.0,
		domain.RiskPhishing:  90.0,
		domain.RiskDebt:      40.0,
		domain.RiskSpam:      20.0,
		domain.RiskSales:     10.0,
		domain.RiskAutoBlock: 0.0,
	}

	var totalRawScore float64
	var lastHumanActivity time.Time
	var autoBlockCount int

	uniqueReporters := make(map[string]bool)

	countryCode := history[0].CountryCode
	now := time.Now().UTC()

	for _, r := range history {
		if r.Category == domain.RiskAutoBlock {
			if now.Sub(r.CreatedAt).Hours() < 24*7 {
				autoBlockCount++
			}
			continue
		}

		uniqueReporters[r.ReporterHash] = true

		if r.CreatedAt.After(lastHumanActivity) {
			lastHumanActivity = r.CreatedAt
		}

		weight := weights[r.Category]
		elapsedDays := now.Sub(r.CreatedAt).Hours() / 24.0
		if elapsedDays < 0 {
			elapsedDays = 0
		}

		decayFactor := math.Pow(0.5, elapsedDays/HalfLifeDays)
		totalRawScore += weight * decayFactor
	}

	reportersCount := len(uniqueReporters)
	var consensusFactor float64

	switch {
	case reportersCount == 1:
		consensusFactor = 0.10
	case reportersCount == 2:
		consensusFactor = 0.20
	case reportersCount == 3:
		consensusFactor = 0.30
	case reportersCount == 4:
		consensusFactor = 0.50
	case reportersCount == 5:
		consensusFactor = 0.70
	default:
		consensusFactor = 1.00
	}

	finalScore := totalRawScore * consensusFactor

	effectiveLastActivity := lastHumanActivity
	if autoBlockCount > 10 {
		effectiveLastActivity = now
		if finalScore < 25 {
			finalScore = 25.0
		}
	}

	if finalScore > 100.0 {
		finalScore = 100.0
	}

	var level domain.RiskLevel
	switch {
	case finalScore >= 60:
		level = domain.LevelCritical
	case finalScore >= 20:
		level = domain.LevelWarning
	default:
		level = domain.LevelSafe
	}

	if finalScore < 5.0 {
		return s.repo.DeleteScore(ctx, phoneNumber, countryCode)
	}

	newScore := &domain.PhoneScore{
		PhoneNumber:      phoneNumber,
		CountryCode:      countryCode,
		Score:            math.Round(finalScore*100) / 100,
		RiskLevel:        level,
		LastActivity:     effectiveLastActivity,
		VelocityHitCount: autoBlockCount,
		TotalReports:     len(history),
	}

	if err := s.repo.UpsertScore(ctx, newScore, OneYearSeconds); err != nil {
		return err
	}
	return s.repo.UpsertCountryThreat(ctx, newScore, OneYearSeconds)
}
