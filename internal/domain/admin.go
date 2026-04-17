package domain

import (
	"time"

	"github.com/lib/pq"
)

type AdminDepartment struct {
	ID          uint           `gorm:"primaryKey;column:id"`
	Name        string         `gorm:"unique;column:name"`
	Description *string        `gorm:"column:description"`
	IsActive    bool           `gorm:"column:isActive;default:true"`
	CreatedAt   time.Time      `gorm:"column:createdAt;default:now()"`
	UpdatedAt   time.Time      `gorm:"column:updatedAt"`
	Permissions pq.StringArray `gorm:"type:text[];column:permissions"`

	Admins []Admin `gorm:"foreignKey:DepartmentID"`
}

func (AdminDepartment) TableName() string {
	return "AdminDepartment"
}

type Admin struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	Email        string    `gorm:"unique;column:email"`
	Name         string    `gorm:"column:name"`
	PhoneNumber  string    `gorm:"column:phoneNumber"`
	PasswordHash string    `gorm:"column:passwordHash"`
	DepartmentID uint      `gorm:"column:departmentId"`
	CreatedAt    time.Time `gorm:"column:createdAt;default:now()"`
	UpdatedAt    time.Time `gorm:"column:updatedAt"`

	Department AdminDepartment `gorm:"foreignKey:DepartmentID;constraint:OnDelete:RESTRICT;"`
}

func (Admin) TableName() string {
	return "Admin"
}
