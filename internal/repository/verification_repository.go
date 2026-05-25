package repository

import (
	"errors"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type verificationRepository struct {
	db *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) domain.AdminVerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) GetPaginated(query domain.VerificationQuery) (*domain.PaginatedVerificationResponse, error) {
	var infos []domain.UserInformation
	var total int64

	db := r.db.Model(&domain.UserInformation{})

	if query.VerificationStatus != "" {
		db = db.Where("verification_status = ?", query.VerificationStatus)
	}

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR last_name ILIKE ? OR phone ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Order("created_date DESC").
		Offset(offset).Limit(query.PageSize).
		Find(&infos).Error
	if err != nil {
		return nil, err
	}

	items := make([]domain.UserVerificationItem, len(infos))
	for i, info := range infos {
		idNum := info.IdentityNumberEncrypted
		if strings.HasPrefix(idNum, "ENC_") {
			idNum = idNum[4:]
		}
		laser := info.LaserIdEncrypted
		if strings.HasPrefix(laser, "ENC_") {
			laser = laser[4:]
		}

		items[i] = domain.UserVerificationItem{
			UserID:                    info.UserId,
			Prefix:                    info.Prefix,
			Name:                      info.Name,
			LastName:                  info.LastName,
			Phone:                     info.Phone,
			Email:                     info.Email,
			IdentityNumber:            idNum,
			LaserID:                   laser,
			IdCardType:                info.IdCardType,
			IdCardPhotoUrl:            info.IdCardPhotoUrl,
			IdCardExpiry:              info.IdCardExpiry,
			VerificationStatus:        info.VerificationStatus,
			VerifiedDate:              info.VerifiedDate,
			RejectionReason:           info.RejectionReason,
			CreatedDate:               info.CreatedDate,
			Birthday:                  info.Birthday,
			HouseNumber:               info.HouseNumber,
			VillageNumber:             info.VillageNumber,
			Alley:                     info.Alley,
			Intersection:              info.Intersection,
			Road:                      info.Road,
			Subdistrict:               info.Subdistrict,
			District:                  info.District,
			Province:                  info.Province,
			PostalCode:                info.PostalCode,
			BuildingName:              info.BuildingName,
			RoomNumber:                info.RoomNumber,
			IsWasteFeeReceipt:         info.IsWasteFeeReceipt,
			IsOnlineTaxPaymentFile:    info.IsOnlineTaxPaymentFile,
			IsOnlineTaxPaymentReceipt: info.IsOnlineTaxPaymentReceipt,
		}
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	return &domain.PaginatedVerificationResponse{
		Items:      items,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: totalPages,
	}, nil
}

func (r *verificationRepository) GetByID(userID uuid.UUID) (*domain.UserVerificationItem, error) {
	var info domain.UserInformation
	err := r.db.Where("user_id = ?", userID).First(&info).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	idNum := info.IdentityNumberEncrypted
	if strings.HasPrefix(idNum, "ENC_") {
		idNum = idNum[4:]
	}
	laser := info.LaserIdEncrypted
	if strings.HasPrefix(laser, "ENC_") {
		laser = laser[4:]
	}

	return &domain.UserVerificationItem{
		UserID:                    info.UserId,
		Prefix:                    info.Prefix,
		Name:                      info.Name,
		LastName:                  info.LastName,
		Phone:                     info.Phone,
		Email:                     info.Email,
		IdentityNumber:            idNum,
		LaserID:                   laser,
		IdCardType:                info.IdCardType,
		IdCardPhotoUrl:            info.IdCardPhotoUrl,
		IdCardExpiry:              info.IdCardExpiry,
		VerificationStatus:        info.VerificationStatus,
		VerifiedDate:              info.VerifiedDate,
		RejectionReason:           info.RejectionReason,
		CreatedDate:               info.CreatedDate,
		Birthday:                  info.Birthday,
		HouseNumber:               info.HouseNumber,
		VillageNumber:             info.VillageNumber,
		Alley:                     info.Alley,
		Intersection:              info.Intersection,
		Road:                      info.Road,
		Subdistrict:               info.Subdistrict,
		District:                  info.District,
		Province:                  info.Province,
		PostalCode:                info.PostalCode,
		BuildingName:              info.BuildingName,
		RoomNumber:                info.RoomNumber,
		IsWasteFeeReceipt:         info.IsWasteFeeReceipt,
		IsOnlineTaxPaymentFile:    info.IsOnlineTaxPaymentFile,
		IsOnlineTaxPaymentReceipt: info.IsOnlineTaxPaymentReceipt,
	}, nil
}

func (r *verificationRepository) Approve(req *domain.ApproveVerificationRequest, adminUserID string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"verification_status":           string(domain.VerificationStatusVerified),
		"verified_date":                 &now,
		"rejection_reason":              nil,
		"updated_by":                    adminUserID,
		"updated_date":                  now,
		"prefix":                        req.Prefix,
		"name":                          req.Name,
		"last_name":                     req.LastName,
		"phone":                         req.Phone,
		"email":                         req.Email,
		"birthday":                      req.Birthday,
		"house_number":                  req.HouseNumber,
		"village_number":                req.VillageNumber,
		"alley":                         req.Alley,
		"intersection":                  req.Intersection,
		"road":                          req.Road,
		"subdistrict":                   req.Subdistrict,
		"district":                      req.District,
		"province":                      req.Province,
		"postal_code":                   req.PostalCode,
		"building_name":                 req.BuildingName,
		"room_number":                   req.RoomNumber,
		"is_waste_fee_receipt":          req.IsWasteFeeReceipt,
		"is_online_tax_payment_file":    req.IsOnlineTaxPaymentFile,
		"is_online_tax_payment_receipt": req.IsOnlineTaxPaymentReceipt,
	}
	return r.db.Model(&domain.UserInformation{}).
		Where("user_id = ?", req.UserID).
		Updates(updates).Error
}

func (r *verificationRepository) Reject(userID uuid.UUID, reason string, adminUserID string) error {
	now := time.Now()
	return r.db.Model(&domain.UserInformation{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"verification_status": string(domain.VerificationStatusRejected),
			"rejection_reason":    &reason,
			"updated_by":          adminUserID,
			"updated_date":        now,
		}).Error
}

func (r *verificationRepository) GetFCMTokens(userID uuid.UUID) ([]string, error) {
	var tokens []domain.UserFCMToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, len(tokens))
	for i, t := range tokens {
		result[i] = t.Token
	}
	return result, nil
}

func (r *verificationRepository) CreateNotification(notification *domain.ModuleNotification) error {
	return r.db.Create(notification).Error
}

func (r *verificationRepository) GetRegisterModuleID() (*uuid.UUID, error) {
	var m domain.Module
	err := r.db.Select("id").Where("key = ? OR is_used_for_user_registration_only = true", "register").Limit(1).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
