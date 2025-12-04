package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rgdevment/spam-registry/internal/domain"
	"github.com/rgdevment/spam-registry/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRepo struct {
	reports []*domain.Report
	scores  map[string]*domain.PhoneScore
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		reports: []*domain.Report{},
		scores:  make(map[string]*domain.PhoneScore),
	}
}

func (m *MockRepo) SaveRawReport(ctx context.Context, r *domain.Report) error {
	m.reports = append(m.reports, r)
	return nil
}

func (m *MockRepo) GetRawReports(ctx context.Context, phone string) ([]*domain.Report, error) {
	var result []*domain.Report
	for _, r := range m.reports {
		if r.PhoneNumber == phone {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRepo) UpsertScore(ctx context.Context, s *domain.PhoneScore, ttl int) error {
	m.scores[s.PhoneNumber] = s
	return nil
}

func (m *MockRepo) UpsertCountryThreat(ctx context.Context, s *domain.PhoneScore, ttl int) error {
	return nil
}

func (m *MockRepo) DeleteScore(ctx context.Context, phone, country string) error {
	delete(m.scores, phone)
	return nil
}

func (m *MockRepo) GetScore(ctx context.Context, phone string) (*domain.PhoneScore, error) {
	if s, exists := m.scores[phone]; exists {
		return s, nil
	}
	return nil, nil
}

func TestQuantumRiskAlgorithm(t *testing.T) {
	cases := []struct {
		Name        string
		TargetPhone string
		Actions     []struct {
			ReporterID string
			Category   domain.RiskCategory
			TimeAgo    time.Duration
		}
		ExpectedLevel domain.RiskLevel
		ExpectedMin   float64
		ExpectedMax   float64
		ShouldExist   bool
	}{
		{
			Name:        "1. El Ex-Rencoroso (1 reporte Fraude reciente)",
			TargetPhone: "+56911111111",
			Actions: []struct {
				ReporterID string
				Category   domain.RiskCategory
				TimeAgo    time.Duration
			}{
				{"user_A", domain.RiskFraud, 0},
			},
			ExpectedLevel: domain.LevelSafe,
			ExpectedMin:   9.0, ExpectedMax: 11.0,
			ShouldExist: true,
		},
		{
			Name:        "2. Estafa Confirmada (3 víctimas recientes)",
			TargetPhone: "+56922222222",
			Actions: []struct {
				ReporterID string
				Category   domain.RiskCategory
				TimeAgo    time.Duration
			}{
				{"user_A", domain.RiskFraud, 0},
				{"user_B", domain.RiskFraud, 0},
				{"user_C", domain.RiskFraud, 0},
			},
			ExpectedLevel: domain.LevelCritical,
			ExpectedMin:   89.0, ExpectedMax: 91.0,
			ShouldExist: true,
		},
		{
			Name:        "3. Decaimiento Natural (Fraude masivo hace 6 meses)",
			TargetPhone: "+56933333333",
			Actions: []struct {
				ReporterID string
				Category   domain.RiskCategory
				TimeAgo    time.Duration
			}{
				{"u1", domain.RiskFraud, 24 * 180 * time.Hour},
				{"u2", domain.RiskFraud, 24 * 180 * time.Hour},
				{"u3", domain.RiskFraud, 24 * 180 * time.Hour},
				{"u4", domain.RiskFraud, 24 * 180 * time.Hour},
				{"u5", domain.RiskFraud, 24 * 180 * time.Hour},
				{"u6", domain.RiskFraud, 24 * 180 * time.Hour},
			},
			ExpectedLevel: domain.LevelCritical,
			ExpectedMin:   99.0, ExpectedMax: 100.0,
			ShouldExist: true,
		},
		{
			Name:        "4. Muerte por Olvido (Spam hace 1 año)",
			TargetPhone: "+56944444444",
			Actions: []struct {
				ReporterID string
				Category   domain.RiskCategory
				TimeAgo    time.Duration
			}{
				{"u1", domain.RiskSpam, 24 * 365 * time.Hour},
				{"u2", domain.RiskSpam, 24 * 365 * time.Hour},
			},
			ExpectedLevel: domain.LevelSafe,
			ExpectedMin:   0.0, ExpectedMax: 0.0,
			ShouldExist: false, // Debe ser borrado
		},
		{
			Name:        "5. El Enjambre Zombie (Auto-Blocks mantienen vivo el ataque)",
			TargetPhone: "+56955555555",
			Actions: []struct {
				ReporterID string
				Category   domain.RiskCategory
				TimeAgo    time.Duration
			}{
				{"u1", domain.RiskFraud, 24 * 365 * time.Hour},

				{"sys", domain.RiskAutoBlock, 0}, {"sys", domain.RiskAutoBlock, 0},
				{"sys", domain.RiskAutoBlock, 0}, {"sys", domain.RiskAutoBlock, 0},
				{"sys", domain.RiskAutoBlock, 0}, {"sys", domain.RiskAutoBlock, 0},
				{"sys", domain.RiskAutoBlock, 0}, {"sys", domain.RiskAutoBlock, 0},
				{"sys", domain.RiskAutoBlock, 0}, {"sys", domain.RiskAutoBlock, 0},
				{"sys", domain.RiskAutoBlock, 0}, {"sys", domain.RiskAutoBlock, 0},
			},
			ExpectedLevel: domain.LevelWarning,
			ExpectedMin:   25.0, ExpectedMax: 25.0,
			ShouldExist: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			repo := NewMockRepo()
			svc := service.NewReportService(repo, "secret_salt")

			for _, action := range tc.Actions {
				report := domain.NewReport(
					tc.TargetPhone,
					"CL",
					action.ReporterID,
					action.Category,
					"Test comment",
				)
				report.CreatedAt = time.Now().UTC().Add(-action.TimeAgo)
				repo.SaveRawReport(context.Background(), report)
			}

			err := svc.CalculateAndSaveRisk(context.Background(), tc.TargetPhone)
			require.NoError(t, err)

			savedScore, _ := repo.GetScore(context.Background(), tc.TargetPhone)

			if !tc.ShouldExist {
				if savedScore != nil {
					assert.Fail(t, "El score debería haber sido borrado (< 5.0) y sigue existiendo")
				}
			} else {
				require.NotNil(t, savedScore, "El score debería existir en la DB")
				fmt.Printf("   -> %s Score Calculado: %.2f (Nivel: %s)\n", tc.Name, savedScore.Score, savedScore.RiskLevel)

				assert.GreaterOrEqual(t, savedScore.Score, tc.ExpectedMin, "Score muy bajo")
				assert.LessOrEqual(t, savedScore.Score, tc.ExpectedMax, "Score muy alto")
				assert.Equal(t, tc.ExpectedLevel, savedScore.RiskLevel, "Nivel de riesgo incorrecto")
			}
		})
	}
}
