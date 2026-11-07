package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/crypto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func FormatAndMaskCitizenID(id string) string {
	digits := ""
	for _, ch := range id {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}
	if len(digits) != 13 {
		return id
	}
	return fmt.Sprintf("%s-%s-••••••-%s-%s", digits[0:1], digits[1:5], digits[10:12], digits[12:13])
}

func FormatFullCitizenID(id string) string {
	digits := ""
	for _, ch := range id {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}
	if len(digits) != 13 {
		return id
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", digits[0:1], digits[1:5], digits[5:10], digits[10:12], digits[12:13])
}

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
		idNum := crypto.DecryptPID(info.IdentityNumberEncrypted)
		laser := crypto.DecryptPID(info.LaserIdEncrypted)
		maskedID := FormatAndMaskCitizenID(idNum)

		cardAddr, resAddr, popType := r.resolveAddressAndPopulation(&info)

		items[i] = domain.UserVerificationItem{
			UserID:                    info.UserId,
			Prefix:                    info.Prefix,
			Name:                      info.Name,
			LastName:                  info.LastName,
			Phone:                     info.Phone,
			Email:                     info.Email,
			IdentityNumber:            maskedID,
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
			CardAddress:               cardAddr,
			ResidentialAddress:        resAddr,
			PopulationType:            popType,
			VerifiedByType:            func() string { if info.VerifiedByType != "" { return info.VerifiedByType }; return "admin" }(),
			IsMasked:                  true,
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

func (r *verificationRepository) resolveAddressAndPopulation(info *domain.UserInformation) (string, string, string) {
	// 1. Build residential address from user_informations
	resParts := []string{}
	if info.HouseNumber != "" {
		resParts = append(resParts, "บ้านเลขที่ "+info.HouseNumber)
	}
	if info.VillageNumber != "" {
		resParts = append(resParts, "หมู่ "+info.VillageNumber)
	}
	if info.Alley != "" {
		resParts = append(resParts, "ซอย "+info.Alley)
	}
	if info.Road != "" {
		resParts = append(resParts, "ถนน "+info.Road)
	}
	if info.Subdistrict != "" {
		resParts = append(resParts, "ต./แขวง "+info.Subdistrict)
	}
	if info.District != "" {
		resParts = append(resParts, "อ./เขต "+info.District)
	}
	if info.Province != "" {
		resParts = append(resParts, "จ."+info.Province)
	}
	if info.PostalCode > 0 {
		resParts = append(resParts, fmt.Sprintf("%d", info.PostalCode))
	}
	residentialAddress := strings.Join(resParts, " ")
	if residentialAddress == "" {
		residentialAddress = "ไม่ระบุที่อยู่อาศัยปัจจุบัน"
	}

	// 2. Query Card/Registered address from user_oauth_accounts (ThaID)
	var oauthAcc domain.UserOauthAccount
	var cardAddress string
	if err := r.db.Where("user_id = ? AND provider ILIKE ?", info.UserId, "%thaiid%").First(&oauthAcc).Error; err == nil {
		if oauthAcc.RawData != "" {
			var rawMap map[string]interface{}
			if json.Unmarshal([]byte(oauthAcc.RawData), &rawMap) == nil {
				if addrObj, ok := rawMap["address"].(map[string]interface{}); ok {
					if formatted, ok := addrObj["formatted"].(string); ok && formatted != "" {
						cardAddress = formatted
					}
				} else if addrStr, ok := rawMap["address"].(string); ok && addrStr != "" {
					cardAddress = addrStr
				}
			}
		}
	}

	if cardAddress == "" {
		cardAddress = residentialAddress
	}

	// 3. Population classification:
	// - บัตร/ทะเบียนบ้านอยู่ชลบุรี -> ประชากรจริง
	// - บัตรเป็นที่อื่น แต่ที่อยู่อาศัยเป็นชลบุรี -> ประชากรแฝง
	// - อื่นๆ -> ประชากรนอกพื้นที่
	isCardChonburi := strings.Contains(cardAddress, "ชลบุรี")
	isResChonburi := strings.Contains(info.Province, "ชลบุรี") || strings.Contains(residentialAddress, "ชลบุรี")

	populationType := "outside"
	if isCardChonburi {
		populationType = "real"
	} else if isResChonburi {
		populationType = "latent"
	}

	return cardAddress, residentialAddress, populationType
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

	idNum := crypto.DecryptPID(info.IdentityNumberEncrypted)
	laser := crypto.DecryptPID(info.LaserIdEncrypted)
	maskedID := FormatAndMaskCitizenID(idNum)

	cardAddr, resAddr, popType := r.resolveAddressAndPopulation(&info)

	return &domain.UserVerificationItem{
		UserID:                    info.UserId,
		Prefix:                    info.Prefix,
		Name:                      info.Name,
		LastName:                  info.LastName,
		Phone:                     info.Phone,
		Email:                     info.Email,
		IdentityNumber:            maskedID,
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
		CardAddress:               cardAddr,
		ResidentialAddress:        resAddr,
		PopulationType:            popType,
		VerifiedByType:            func() string { if info.VerifiedByType != "" { return info.VerifiedByType }; return "admin" }(),
		IsMasked:                  true,
	}, nil
}

func (r *verificationRepository) GetUnmaskedCitizenID(userID uuid.UUID) (string, error) {
	var info domain.UserInformation
	err := r.db.Select("identity_number_encrypted").Where("user_id = ?", userID).First(&info).Error
	if err != nil {
		return "", err
	}
	idNum := crypto.DecryptPID(info.IdentityNumberEncrypted)
	return FormatFullCitizenID(idNum), nil
}

func (r *verificationRepository) Approve(req *domain.ApproveVerificationRequest, adminUserID string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"verification_status":           string(domain.VerificationStatusVerified),
		"verified_date":                 &now,
		"verified_by_type":              "admin",
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

	if req.IdentityNumber != "" {
		cleanID := crypto.ExtractDigits(req.IdentityNumber)
		if len(cleanID) == 13 {
			encID, err := crypto.EncryptPID(cleanID)
			if err == nil {
				updates["identity_number_encrypted"] = encID
				updates["identity_number_hash"] = crypto.HashPID(cleanID)
			}
		}
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
	err := r.db.Select("id").Where("key = ? OR key = ?", "register", "ModuleIdentityVerifications").Limit(1).First(&m).Error
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
