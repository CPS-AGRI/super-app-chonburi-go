package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestResolveNotificationModuleID(t *testing.T) {
	fallback := moduleIDComplaintCenter

	assert.Equal(t, uuid.MustParse("eadaf67c-db3e-4557-82cd-8bf721c3e2e7"), resolveNotificationModuleID("tax", fallback))
	assert.Equal(t, uuid.MustParse("5b630777-4de8-42f7-926a-2335879e0f6d"), resolveNotificationModuleID("complaint", fallback))
	assert.Equal(t, uuid.MustParse("d01b2ce5-34a9-498b-bba0-b1b8360f1ea9"), resolveNotificationModuleID("complaint_center", fallback))
	assert.Equal(t, uuid.MustParse("8c7ce421-5d1f-41de-840b-14ac192d4778"), resolveNotificationModuleID("user_verify", fallback))
	assert.Equal(t, uuid.MustParse("d04d6d51-b35d-43df-be91-31178986aae3"), resolveNotificationModuleID("pr", fallback))
	assert.Equal(t, uuid.MustParse("2c8e1925-5d1f-41de-840b-14ac192d4778"), resolveNotificationModuleID("cctv", fallback))
	assert.Equal(t, uuid.MustParse("10e5e192-8254-479b-ab83-86bfd09504eb"), resolveNotificationModuleID("weather", fallback))
	assert.Equal(t, fallback, resolveNotificationModuleID("unknown_module", fallback))
}

func TestBuildDepartmentNotification_UsesExpectedModuleID(t *testing.T) {
	taxNotif := buildDepartmentNotification("tax", "", "Employees", "มีรายการยื่นแบบภาษีใหม่", "body", "ref-1", "pending")
	assert.Equal(t, uuid.MustParse("eadaf67c-db3e-4557-82cd-8bf721c3e2e7"), taxNotif.ModuleID)
	assert.Equal(t, "admin", taxNotif.Type)
	assert.NotNil(t, taxNotif.Role)
	assert.Equal(t, "Employees", *taxNotif.Role)

	centerNotif := buildDepartmentNotification("complaint_center", "", "Managers", "เรื่องร้องเรียน", "body", "ref-2", "received")
	assert.Equal(t, uuid.MustParse("d01b2ce5-34a9-498b-bba0-b1b8360f1ea9"), centerNotif.ModuleID)
}

