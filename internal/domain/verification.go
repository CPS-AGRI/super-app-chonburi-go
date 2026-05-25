package domain

import (
	"time"

	"github.com/google/uuid"
)

type VerificationStatus string

const (
	VerificationStatusUnverified VerificationStatus = "unverified"
	VerificationStatusPending    VerificationStatus = "pending"
	VerificationStatusVerified   VerificationStatus = "verified"
	VerificationStatusRejected   VerificationStatus = "rejected"
)

type VerificationQuery struct {
	PageNumber         int    `json:"page_number"`
	PageSize           int    `json:"page_size"`
	Search             string `json:"search"`
	VerificationStatus string `json:"verification_status"`
}

type UserVerificationItem struct {
	UserID                    uuid.UUID  `json:"user_id"`
	Prefix                    string     `json:"prefix"`
	Name                      string     `json:"name"`
	LastName                  string     `json:"last_name"`
	Phone                     string     `json:"phone"`
	Email                     *string    `json:"email"`
	IdentityNumber            string     `json:"identity_number"` // Decrypted in code or raw
	LaserID                   string     `json:"laser_id"`        // Decrypted in code or raw
	IdCardType                *int       `json:"id_card_type"`
	IdCardPhotoUrl            *string    `json:"id_card_photo_url"`
	IdCardExpiry              *time.Time `json:"id_card_expiry"`
	VerificationStatus        string     `json:"verification_status"`
	VerifiedDate              *time.Time `json:"verified_date"`
	RejectionReason           *string    `json:"rejection_reason"`
	CreatedDate               time.Time  `json:"created_date"`
	Birthday                  *time.Time `json:"birthday"`
	HouseNumber               string     `json:"house_number"`
	VillageNumber             string     `json:"village_number"`
	Alley                     string     `json:"alley"`
	Intersection              string     `json:"intersection"`
	Road                      string     `json:"road"`
	Subdistrict               string     `json:"subdistrict"`
	District                  string     `json:"district"`
	Province                  string     `json:"province"`
	PostalCode                int        `json:"postal_code"`
	BuildingName              string     `json:"building_name"`
	RoomNumber                string     `json:"room_number"`
	IsWasteFeeReceipt         bool       `json:"is_waste_fee_receipt"`
	IsOnlineTaxPaymentFile    bool       `json:"is_online_tax_payment_file"`
	IsOnlineTaxPaymentReceipt bool       `json:"is_online_tax_payment_receipt"`
}

type PaginatedVerificationResponse struct {
	Items      []UserVerificationItem `json:"items"`
	TotalItems int64                  `json:"total_items"`
	PageNumber int                    `json:"page_number"`
	TotalPages int                    `json:"total_pages"`
}

type ApproveVerificationRequest struct {
	UserID                    uuid.UUID  `json:"user_id" validate:"required"`
	Prefix                    string     `json:"prefix"`
	Name                      string     `json:"name"`
	LastName                  string     `json:"last_name"`
	Phone                     string     `json:"phone"`
	Email                     *string    `json:"email"`
	Birthday                  *time.Time `json:"birthday"`
	HouseNumber               string     `json:"house_number"`
	VillageNumber             string     `json:"village_number"`
	Alley                     string     `json:"alley"`
	Intersection              string     `json:"intersection"`
	Road                      string     `json:"road"`
	Subdistrict               string     `json:"subdistrict"`
	District                  string     `json:"district"`
	Province                  string     `json:"province"`
	PostalCode                int        `json:"postal_code"`
	BuildingName              string     `json:"building_name"`
	RoomNumber                string     `json:"room_number"`
	IsWasteFeeReceipt         bool       `json:"is_waste_fee_receipt"`
	IsOnlineTaxPaymentFile    bool       `json:"is_online_tax_payment_file"`
	IsOnlineTaxPaymentReceipt bool       `json:"is_online_tax_payment_receipt"`
}

type RejectVerificationRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Reason string    `json:"reason"  validate:"required"`
}

type AdminVerificationRepository interface {
	GetPaginated(query VerificationQuery) (*PaginatedVerificationResponse, error)
	GetByID(userID uuid.UUID) (*UserVerificationItem, error)
	Approve(req *ApproveVerificationRequest, adminUserID string) error
	Reject(userID uuid.UUID, reason string, adminUserID string) error
	GetFCMTokens(userID uuid.UUID) ([]string, error)
	CreateNotification(notification *ModuleNotification) error
	GetRegisterModuleID() (*uuid.UUID, error)
}

type AdminVerificationUseCase interface {
	GetVerifications(query VerificationQuery) (*PaginatedVerificationResponse, error)
	GetVerificationByID(userID uuid.UUID) (*UserVerificationItem, error)
	ApproveVerification(req *ApproveVerificationRequest, adminUserID string) error
	RejectVerification(userID uuid.UUID, reason string, adminUserID string) error
}
