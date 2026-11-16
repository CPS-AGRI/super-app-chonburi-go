package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type sqlCaptureLogger struct {
	queries *[]string
}

func (l *sqlCaptureLogger) LogMode(level logger.LogLevel) logger.Interface { return l }
func (l *sqlCaptureLogger) Info(ctx context.Context, s string, i ...interface{}) {
}
func (l *sqlCaptureLogger) Warn(ctx context.Context, s string, i ...interface{}) {
}
func (l *sqlCaptureLogger) Error(ctx context.Context, s string, i ...interface{}) {
}
func (l *sqlCaptureLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sqlStr, _ := fc()
	*l.queries = append(*l.queries, sqlStr)
}

func newDryRunKPIRepo(t *testing.T) (*kpiLogRepository, *[]string) {
	t.Helper()
	var queries []string
	db, err := gorm.Open(postgres.New(postgres.Config{
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               &sqlCaptureLogger{queries: &queries},
	})
	require.NoError(t, err)

	return &kpiLogRepository{db: db}, &queries
}

func TestComplaintKPIEventsCTE_StructureAndDeduplication(t *testing.T) {
	cte := complaintKPIEventsCTE(domain.ModuleKeyComplaint)
	assert.Contains(t, cte, "WITH live_complaints AS")
	assert.Contains(t, cte, "FROM module_complaints mc")
	assert.Contains(t, cte, "FROM module_complaint_activities act")
	assert.Contains(t, cte, "COALESCE(lc.assigner_id, lc.asg_actor_id)::uuid AS actor_id")
	assert.Contains(t, cte, "'COMPLAINT_MANAGER_ASSIGNED'")
	assert.Contains(t, cte, "'COMPLAINT_STAFF_ACKNOWLEDGED'")
	assert.Contains(t, cte, "'COMPLAINT_STAFF_RESOLVED'")
	assert.Contains(t, cte, "NOT EXISTS")

	nonComplaintCTE := complaintKPIEventsCTE(domain.ModuleKeyTaxHotel)
	assert.NotContains(t, nonComplaintCTE, "live_complaints")
	assert.Contains(t, nonComplaintCTE, "WITH complaint_kpi_events AS")
}

func TestComplaintKPIQueries_UseUnifiedEventsCTEAndAcknowledgedCode(t *testing.T) {
	repo, queries := newDryRunKPIRepo(t)
	filter := domain.KPIQueryFilter{ModuleKey: domain.ModuleKeyComplaint}

	checkDryRunErr := func(err error) {
		if err != nil && !errors.Is(err, gorm.ErrDryRunModeUnsupported) {
			require.NoError(t, err)
		}
	}

	_, err := repo.GetKPISummary(filter)
	checkDryRunErr(err)
	_, err = repo.GetStaffLeaderboard(filter)
	checkDryRunErr(err)
	_, err = repo.GetManagerScorecard(filter)
	checkDryRunErr(err)
	_, err = repo.GetStaffDetail("11111111-1111-1111-1111-111111111111", filter)
	checkDryRunErr(err)
	_, err = repo.GetManagerDetail("22222222-2222-2222-2222-222222222222", filter)
	checkDryRunErr(err)

	joined := strings.Join(*queries, "\n---\n")
	assert.Contains(t, joined, "FROM complaint_kpi_events k")
	assert.Contains(t, joined, "LEFT JOIN complaint_kpi_events k ON k.actor_id = a.id")
	assert.Contains(t, joined, "COMPLAINT_STAFF_ACKNOWLEDGED")
}
