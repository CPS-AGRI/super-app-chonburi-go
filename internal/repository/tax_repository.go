package repository

import (
	"errors"
	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type taxRepository struct {
	db *gorm.DB
}

func NewTaxRepository(db *gorm.DB) domain.TaxRepository {
	return &taxRepository{db: db}
}

func (r *taxRepository) CreateImport(importHead *domain.ModuleOnlineTaxPayment) error {
	return r.db.Create(importHead).Error
}

func (r *taxRepository) CreateInformation(info *domain.ModuleOnlineTaxPaymentInformation) error {
	return r.db.Create(info).Error
}

func (r *taxRepository) GetImportByYearAndName(year, name string) (*domain.ModuleOnlineTaxPayment, error) {
	var importHead domain.ModuleOnlineTaxPayment
	err := r.db.Where("year = ? AND name = ?", year, name).First(&importHead).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &importHead, nil
}

func (r *taxRepository) GetInformationByRefs(ref1, ref2 string) (*domain.ModuleOnlineTaxPaymentInformation, error) {
	var info domain.ModuleOnlineTaxPaymentInformation
	err := r.db.Where("reference_number_1 = ? AND reference_number_2 = ?", ref1, ref2).First(&info).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

func (r *taxRepository) GetInformationsPaginated(query domain.TaxQuery) (*domain.PaginatedTaxResponse, error) {
	var items []domain.ModuleOnlineTaxPaymentInformation
	var total int64

	db := r.db.Model(&domain.ModuleOnlineTaxPaymentInformation{})

	if query.ModuleTypeId != "" {
		db = db.Where("module_type_id = ?", query.ModuleTypeId)
	}
	if len(query.Status) > 0 {
		db = db.Where("status IN ?", query.Status)
	}
	if len(query.LinkStatus) > 0 {
		db = db.Where("link_status IN ?", query.LinkStatus)
	}
	if query.IdentityNumber != "" {
		db = db.Where("identity_number = ?", query.IdentityNumber)
	}
	if query.Keyword != "" {
		db = db.Where("name ILIKE ? OR identity_number ILIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	db.Count(&total)

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("User").
		Offset(offset).Limit(query.PageSize).
		Order("created_date DESC").
		Find(&items).Error

	return &domain.PaginatedTaxResponse{
		Items:      items,
		TotalItems: total,
		PageNumber: query.PageNumber,
		PageSize:   query.PageSize,
	}, err
}

func (r *taxRepository) GetInformationByID(id string) (*domain.ModuleOnlineTaxPaymentInformation, error) {
	var info domain.ModuleOnlineTaxPaymentInformation
	err := r.db.Preload("User").Where("id = ?", id).First(&info).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

func (r *taxRepository) GetImportsPaginated(query domain.TaxQuery) ([]domain.ModuleOnlineTaxPayment, int64, error) {
	var items []domain.ModuleOnlineTaxPayment
	var total int64

	db := r.db.Model(&domain.ModuleOnlineTaxPayment{})

	if query.Year != "" {
		db = db.Where("year = ?", query.Year)
	}

	db.Count(&total)

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).
		Order("created_date DESC").
		Find(&items).Error

	return items, total, err
}

func (r *taxRepository) UpdateInformation(info *domain.ModuleOnlineTaxPaymentInformation) error {
	return r.db.Save(info).Error
}

func (r *taxRepository) CreateLog(log *domain.ModuleOnlineTaxPaymentLog) error {
	return r.db.Create(log).Error
}

func (r *taxRepository) GetInformationsByIdentityNumber(identityNumber string) ([]domain.ModuleOnlineTaxPaymentInformation, error) {
	var items []domain.ModuleOnlineTaxPaymentInformation
	err := r.db.Where("identity_number = ?", identityNumber).
		Order("created_date DESC").
		Find(&items).Error
	return items, err
}
