package scylla

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/rgdevment/spam-registry/internal/domain"
	"github.com/rgdevment/spam-registry/internal/service"
)

type scyllaRepository struct {
	session *gocql.Session
}

func NewScyllaRepository(session *gocql.Session) service.Repository {
	return &scyllaRepository{
		session: session,
	}
}

func Connect(keyspace string, hosts ...string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.Timeout = 5 * time.Second
	cluster.ConnectTimeout = 5 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to scylla: %w", err)
	}

	log.Println("✅ Connected to ScyllaDB")
	return session, nil
}

func (r *scyllaRepository) SaveRawReport(ctx context.Context, report *domain.Report) error {
	query := `
        INSERT INTO reports (id, phone_number, country_code, reporter_hash, category, comment, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?) USING TTL ?`

	const ttlSeconds = 47304000

	err := r.session.Query(query,
		report.ID.String(),
		report.PhoneNumber,
		report.CountryCode,
		report.ReporterHash,
		string(report.Category),
		report.Comment,
		report.CreatedAt,
		ttlSeconds,
	).WithContext(ctx).Exec()

	if err != nil {
		return fmt.Errorf("scylla: failed to save raw report: %w", err)
	}

	return nil
}

func (r *scyllaRepository) GetRawReports(ctx context.Context, phoneNumber string) ([]*domain.Report, error) {
	query := `SELECT id, phone_number, country_code, reporter_hash, category, comment, created_at 
	          FROM reports WHERE phone_number = ?`

	iter := r.session.Query(query, phoneNumber).WithContext(ctx).Iter()

	var reports []*domain.Report
	var id gocql.UUID
	var phone, country, hash, catStr, comment string
	var createdAt time.Time

	for iter.Scan(&id, &phone, &country, &hash, &catStr, &comment, &createdAt) {
		parsedID, _ := uuid.Parse(id.String())
		reports = append(reports, &domain.Report{
			ID:           parsedID,
			PhoneNumber:  phone,
			CountryCode:  country,
			ReporterHash: hash,
			Category:     domain.RiskCategory(catStr),
			Comment:      comment,
			CreatedAt:    createdAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("scylla: failed to iterate reports: %w", err)
	}

	return reports, nil
}

func (r *scyllaRepository) GetScore(ctx context.Context, phoneNumber string) (*domain.PhoneScore, error) {
	query := `
        SELECT phone_number, country_code, score, risk_level, last_activity, velocity_hit_count, total_reports 
        FROM scores WHERE phone_number = ?`

	var s domain.PhoneScore
	var riskLevelStr string

	err := r.session.Query(query, phoneNumber).WithContext(ctx).Scan(
		&s.PhoneNumber,
		&s.CountryCode,
		&s.Score,
		&riskLevelStr,
		&s.LastActivity,
		&s.VelocityHitCount,
		&s.TotalReports,
	)

	if err == gocql.ErrNotFound {
		return &domain.PhoneScore{
			PhoneNumber: phoneNumber,
			Score:       0,
			RiskLevel:   domain.LevelSafe,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("scylla: failed to get score: %w", err)
	}

	s.RiskLevel = domain.RiskLevel(riskLevelStr)
	return &s, nil
}

func (r *scyllaRepository) UpsertScore(ctx context.Context, s *domain.PhoneScore, ttlSeconds int) error {
	query := `
        UPDATE scores USING TTL ?
        SET score = ?, 
            risk_level = ?, 
            last_activity = ?, 
            velocity_hit_count = ?, 
            total_reports = ?,
            country_code = ? 
        WHERE phone_number = ?`

	return r.session.Query(query,
		ttlSeconds,
		s.Score,
		string(s.RiskLevel),
		s.LastActivity,
		s.VelocityHitCount,
		s.TotalReports,
		s.CountryCode,
		s.PhoneNumber,
	).WithContext(ctx).Exec()
}

func (r *scyllaRepository) UpsertCountryThreat(ctx context.Context, s *domain.PhoneScore, ttlSeconds int) error {
	if s.RiskLevel == domain.LevelSafe {
		return nil
	}

	query := `
        INSERT INTO active_threats (country_code, risk_level, phone_number, score, last_updated)
        VALUES (?, ?, ?, ?, ?) USING TTL ?`

	return r.session.Query(query,
		s.CountryCode,
		string(s.RiskLevel),
		s.PhoneNumber,
		s.Score,
		s.LastActivity,
		ttlSeconds,
	).WithContext(ctx).Exec()
}

func (r *scyllaRepository) DeleteScore(ctx context.Context, phoneNumber string, countryCode string) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)
	batch.Query("DELETE FROM scores WHERE phone_number = ?", phoneNumber)

	return r.session.ExecuteBatch(batch)
}
