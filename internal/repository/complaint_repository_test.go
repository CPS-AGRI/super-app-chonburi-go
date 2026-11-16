package repository

import (
	"strings"
	"testing"

	"super-app-chonburi-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestComplaintRepository_DepartmentScopeIsolatesForwardedComplaints(t *testing.T) {
	t.Run("center user on manage page (scope=department) is scoped strictly to own departments", func(t *testing.T) {
		kpiRepo, queries := newDryRunKPIRepo(t)
		repo := &complaintRepository{db: kpiRepo.db}

		_, _ = repo.GetPaginated(domain.ComplaintQuery{
			PageNumber:           1,
			PageSize:             50,
			Scope:                "department",
			IsComplaintCenter:    true,
			AdminDepartmentIDs:   []string{"dept-dev"},
			AllowedModuleTypeIDs: []string{"type-1"},
			Status:               []string{domain.ComplaintStatusReceived},
		})

		joined := strings.Join(*queries, "\n")
		assert.Contains(t, joined, "module_complaints.department_id IN ('dept-dev')")
	})

	t.Run("center user on complaint-center (scope=center) sees all center queue", func(t *testing.T) {
		kpiRepo, queries := newDryRunKPIRepo(t)
		repo := &complaintRepository{db: kpiRepo.db}

		_, _ = repo.GetPaginated(domain.ComplaintQuery{
			PageNumber:           1,
			PageSize:             50,
			Scope:                "center",
			IsComplaintCenter:    true,
			AdminDepartmentIDs:   []string{"dept-dev"},
			AllowedModuleTypeIDs: []string{"type-1"},
			Status:               []string{domain.ComplaintStatusPending},
		})

		joined := strings.Join(*queries, "\n")
		assert.NotContains(t, joined, "module_complaints.department_id IN ('dept-dev')")
	})
}

