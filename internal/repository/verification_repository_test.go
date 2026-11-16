package repository

import (
	"strings"
	"testing"

	"super-app-chonburi-go/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestApplyActiveUserVerificationScope_NoDeletedAtAndExcludesSoftDeleted(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	require.NoError(t, err)

	var infos []domain.UserInformation
	stmt := applyActiveUserVerificationScope(db.Model(&domain.UserInformation{})).Find(&infos).Statement
	sqlLower := strings.ToLower(stmt.SQL.String())

	assert.NotContains(t, sqlLower, "deleted_at", "users table has no deleted_at column")
	assert.Contains(t, sqlLower, "inner join users u on u.id = user_informations.user_id")
	assert.Contains(t, sqlLower, "u.phone_number")
	assert.Contains(t, sqlLower, "user_informations.status")
	assert.Contains(t, sqlLower, "user_informations.phone")
}

