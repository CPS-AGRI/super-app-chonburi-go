package usecase_test

import (
	"testing"

	"super-app-chonburi-go/internal/usecase"

	"github.com/stretchr/testify/assert"
)

func TestIsCitizenIDUnmaskAllowed(t *testing.T) {
	// Allowed: SuperAdmin and Managers only
	assert.True(t, usecase.IsCitizenIDUnmaskAllowed("SuperAdmin"))
	assert.True(t, usecase.IsCitizenIDUnmaskAllowed("super_admin"))
	assert.True(t, usecase.IsCitizenIDUnmaskAllowed("Managers"))
	assert.True(t, usecase.IsCitizenIDUnmaskAllowed("Supervisor"))
	assert.True(t, usecase.IsCitizenIDUnmaskAllowed("หัวหน้าฝ่ายปฏิบัติการ"))

	// Denied: Employees, Executives, empty/unknown
	assert.False(t, usecase.IsCitizenIDUnmaskAllowed("Employees"))
	assert.False(t, usecase.IsCitizenIDUnmaskAllowed("Executives"))
	assert.False(t, usecase.IsCitizenIDUnmaskAllowed("executive"))
	assert.False(t, usecase.IsCitizenIDUnmaskAllowed("พนักงานปฏิบัติการ"))
	assert.False(t, usecase.IsCitizenIDUnmaskAllowed("ผู้บริหารระดับสูง"))
	assert.False(t, usecase.IsCitizenIDUnmaskAllowed(""))
}

