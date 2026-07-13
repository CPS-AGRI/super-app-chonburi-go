package repository

import (
	"errors"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type taxNewRepository struct {
	db *gorm.DB
}

func NewTaxNewRepository(db *gorm.DB) domain.TaxNewRepository {
	return &taxNewRepository{db: db}
}

func (r *taxNewRepository) GetBusinessByRegNumber(regNumber string) (*domain.TaxBusiness, error) {
	var business domain.TaxBusiness
	err := r.db.Where("business_reg_number = ?", regNumber).First(&business).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &business, nil
}

func (r *taxNewRepository) GetBusinessByOwnerIdentityNumber(identityNumber string) (*domain.TaxBusiness, error) {
	var business domain.TaxBusiness
	err := r.db.Where("owner_identity_number = ?", identityNumber).First(&business).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &business, nil
}

func (r *taxNewRepository) GetActiveTaxRate(taxType string) (*domain.TaxRate, error) {
	var rate domain.TaxRate
	err := r.db.Where("tax_type = ? AND is_active = ?", taxType, true).First(&rate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rate, nil
}

func (r *taxNewRepository) GetLatestDeclarationVersion(regNumber string, taxType string, month, year int) (int, error) {
	var maxVersion int
	err := r.db.Model(&domain.TaxDeclaration{}).
		Where("business_reg_number = ? AND tax_type = ? AND tax_month = ? AND tax_year = ?", regNumber, taxType, month, year).
		Select("COALESCE(MAX(declaration_version), 0)").
		Row().Scan(&maxVersion)
	return maxVersion, err
}

func (r *taxNewRepository) CreateDeclaration(declaration *domain.TaxDeclaration) error {
	return r.db.Create(declaration).Error
}

func (r *taxNewRepository) GetDeclarationByID(id uuid.UUID) (*domain.TaxDeclaration, error) {
	var declaration domain.TaxDeclaration
	err := r.db.Preload("Business").Where("id = ?", id).First(&declaration).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &declaration, nil
}

func (r *taxNewRepository) UpdateDeclaration(declaration *domain.TaxDeclaration) error {
	return r.db.Save(declaration).Error
}

func (r *taxNewRepository) CreateReconciliationBatch(batch *domain.BankReconciliationBatch) error {
	return r.db.Create(batch).Error
}

func (r *taxNewRepository) CreateReconciliationRecord(record *domain.BankReconciliationRecord) error {
	return r.db.Create(record).Error
}

func (r *taxNewRepository) GetDeclarationByRefs(ref1, ref2 string) (*domain.TaxDeclaration, error) {
	var declaration domain.TaxDeclaration

	err := r.db.Preload("Business").Where("TRIM(ref1) = TRIM(?) AND TRIM(ref2) = TRIM(?)", ref1, ref2).First(&declaration).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &declaration, nil
}

func (r *taxNewRepository) GetUnmatchedReconciliationRecords() ([]domain.BankReconciliationRecord, error) {
	var records []domain.BankReconciliationRecord
	err := r.db.Where("is_matched = ?", false).Order("payment_date DESC").Find(&records).Error
	return records, err
}

func (r *taxNewRepository) UpsertElaasSummary(summary *domain.ElaasDailySummary) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "summary_date"}, {Name: "tax_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_amount", "transaction_count", "filename", "uploaded_by", "created_at"}),
	}).Create(summary).Error
}

func (r *taxNewRepository) GetDashboardSummary(startDate, endDate time.Time) (*domain.DashboardSummaryResponse, error) {
	var totalRevenue float64
	var totalTransactions int64

	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	err := r.db.Model(&domain.TaxDeclaration{}).
		Where("payment_status IN ? AND paid_at BETWEEN ? AND ?", []string{"paid", "verified", "audit_failed"}, startOfDay, endOfDay).
		Select("COALESCE(SUM(calculated_tax), 0)").
		Row().Scan(&totalRevenue)
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&domain.TaxDeclaration{}).
		Where("payment_status IN ? AND paid_at BETWEEN ? AND ?", []string{"paid", "verified", "audit_failed"}, startOfDay, endOfDay).
		Count(&totalTransactions).Error
	if err != nil {
		return nil, err
	}

	type BreakdownRow struct {
		TaxType string  `gorm:"column:tax_type"`
		Amount  float64 `gorm:"column:amount"`
		Count   int     `gorm:"column:count"`
	}
	var dbRows []BreakdownRow
	err = r.db.Model(&domain.TaxDeclaration{}).
		Where("payment_status IN ? AND paid_at BETWEEN ? AND ?", []string{"paid", "verified", "audit_failed"}, startOfDay, endOfDay).
		Select("tax_type, COALESCE(SUM(calculated_tax), 0) as amount, COUNT(*) as count").
		Group("tax_type").
		Find(&dbRows).Error
	if err != nil {
		return nil, err
	}

	breakdown := map[string]domain.TaxCategorySummary{
		"hotel_fee":   {Amount: 0, Count: 0},
		"oil_gas_tax": {Amount: 0, Count: 0},
		"tobacco_tax": {Amount: 0, Count: 0},
	}
	for _, row := range dbRows {
		breakdown[row.TaxType] = domain.TaxCategorySummary{
			Amount: row.Amount,
			Count:  row.Count,
		}
	}

	type TrendRow struct {
		DateStr string  `gorm:"column:date_str"`
		TaxType string  `gorm:"column:tax_type"`
		Amount  float64 `gorm:"column:amount"`
	}
	var trendRows []TrendRow

	err = r.db.Model(&domain.TaxDeclaration{}).
		Where("payment_status IN ? AND paid_at BETWEEN ? AND ?", []string{"paid", "verified", "audit_failed"}, startOfDay, endOfDay).
		Select("TO_CHAR(paid_at, 'YYYY-MM-DD') as date_str, tax_type, COALESCE(SUM(calculated_tax), 0) as amount").
		Group("TO_CHAR(paid_at, 'YYYY-MM-DD'), tax_type").
		Order("date_str ASC").
		Find(&trendRows).Error
	if err != nil {
		return nil, err
	}

	trendMap := make(map[string]*domain.DailyTrendItem)
	for _, row := range trendRows {
		if _, ok := trendMap[row.DateStr]; !ok {
			trendMap[row.DateStr] = &domain.DailyTrendItem{
				Date: row.DateStr,
			}
		}
		item := trendMap[row.DateStr]
		switch row.TaxType {
		case "hotel_fee":
			item.HotelFee = row.Amount
		case "oil_gas_tax":
			item.OilGasTax = row.Amount
		case "tobacco_tax":
			item.TobaccoTax = row.Amount
		}
	}

	var dailyTrends []domain.DailyTrendItem
	for d := startOfDay; !d.After(endOfDay); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if item, ok := trendMap[dateStr]; ok {
			dailyTrends = append(dailyTrends, *item)
		} else {
			dailyTrends = append(dailyTrends, domain.DailyTrendItem{
				Date: dateStr,
			})
		}
	}

	return &domain.DashboardSummaryResponse{
		TotalRevenue:      totalRevenue,
		TotalTransactions: int(totalTransactions),
		Breakdown:         breakdown,
		DailyTrends:       dailyTrends,
	}, nil
}

func (r *taxNewRepository) UpsertBusiness(business *domain.TaxBusiness) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "business_reg_number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name_th",
			"name_en",
			"tax_type",
			"owner_name",
			"owner_identity_number",
			"contact_phone",
			"contact_email",
			"address_detail",
			"updated_at",
		}),
	}).Create(business).Error
}

func (r *taxNewRepository) ListDeclarations(taxType, status, search string, startDate, endDate *time.Time, limit, offset int) ([]domain.TaxDeclaration, int64, error) {
	var declarations []domain.TaxDeclaration
	var total int64

	query := r.db.Model(&domain.TaxDeclaration{}).Preload("Business")

	if taxType != "" {
		query = query.Where("tax_declarations.tax_type = ?", taxType)
	}

	if status != "" {
		query = query.Where("tax_declarations.payment_status = ?", status)
	}

	if startDate != nil {
		startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		query = query.Where("tax_declarations.created_at >= ?", startOfDay)
	}

	if endDate != nil {
		endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
		query = query.Where("tax_declarations.created_at <= ?", endOfDay)
	}

	if search != "" {
		query = query.Joins("JOIN tax_businesses ON tax_businesses.id = tax_declarations.business_id").
			Where("tax_declarations.business_reg_number LIKE ? OR tax_businesses.name_th LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if search != "" {
		query = query.Select("tax_declarations.*")
	}

	err = query.Order("tax_declarations.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&declarations).Error
	if err != nil {
		return nil, 0, err
	}

	return declarations, total, nil
}

func (r *taxNewRepository) GetUserInformationByPhoneOrEmail(phone *string, email string) (*domain.UserInformation, error) {
	var userInfo domain.UserInformation
	err := r.db.Where("phone = ? OR email = ?", phone, email).First(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userInfo, nil
}
