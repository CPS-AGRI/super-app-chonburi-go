package repository

import (
	"testing"

	"super-app-chonburi-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPublicRelationCommentPreloadAndModerationErrors(t *testing.T) {
	assert.EqualError(t, domain.ErrPRCommentNotFound, "comment not found")

	repo, queries := newDryRunKPIRepo(t)
	prRepo := &publicRelationRepository{db: repo.db}

	_, _ = prRepo.GetPaginated("d04d6d51-b35d-43df-be91-31178986aae3", domain.PublicRelationQuery{
		PageNumber: 1,
		PageSize:   10,
	})
	_, _ = prRepo.GetByID("d04d6d51-b35d-43df-be91-31178986aae3", "11111111-1111-1111-1111-111111111111")

	assert.NotEmpty(t, *queries)
}

