package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendHTML(to []string, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

type MockTaxNewRepository struct {
	mock.Mock
}

func (m *MockTaxNewRepository) GetBusinessByRegNumber(regNumber string) (*domain.TaxBusiness, error) {
	args := m.Called(regNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaxBusiness), args.Error(1)
}

func (m *MockTaxNewRepository) GetBusinessByOwnerIdentityNumber(identityNumber string) (*domain.TaxBusiness, error) {
	args := m.Called(identityNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaxBusiness), args.Error(1)
}

func (m *MockTaxNewRepository) GetActiveTaxRate(taxType string) (*domain.TaxRate, error) {
	args := m.Called(taxType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaxRate), args.Error(1)
}

func (m *MockTaxNewRepository) GetLatestDeclarationVersion(regNumber string, taxType string, month, year int) (int, error) {
	args := m.Called(regNumber, taxType, month, year)
	return args.Int(0), args.Error(1)
}

func (m *MockTaxNewRepository) CreateDeclaration(declaration *domain.TaxDeclaration) error {
	args := m.Called(declaration)
	return args.Error(0)
}

func (m *MockTaxNewRepository) GetDeclarationByID(id uuid.UUID) (*domain.TaxDeclaration, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaxDeclaration), args.Error(1)
}

func (m *MockTaxNewRepository) UpdateDeclaration(declaration *domain.TaxDeclaration) error {
	args := m.Called(declaration)
	return args.Error(0)
}

func (m *MockTaxNewRepository) ListDeclarations(taxType, status, search string, startDate, endDate *time.Time, limit, offset int) ([]domain.TaxDeclaration, int64, error) {
	args := m.Called(taxType, status, search, startDate, endDate, limit, offset)
	return args.Get(0).([]domain.TaxDeclaration), int64(args.Int(1)), args.Error(2)
}

func (m *MockTaxNewRepository) CreateReconciliationBatch(batch *domain.BankReconciliationBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}

func (m *MockTaxNewRepository) CreateReconciliationRecord(record *domain.BankReconciliationRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockTaxNewRepository) GetDeclarationByRefs(ref1, ref2 string) (*domain.TaxDeclaration, error) {
	args := m.Called(ref1, ref2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaxDeclaration), args.Error(1)
}

func (m *MockTaxNewRepository) GetUnmatchedReconciliationRecords() ([]domain.BankReconciliationRecord, error) {
	args := m.Called()
	return args.Get(0).([]domain.BankReconciliationRecord), args.Error(1)
}

func (m *MockTaxNewRepository) UpsertElaasSummary(summary *domain.ElaasDailySummary) error {
	args := m.Called(summary)
	return args.Error(0)
}

func (m *MockTaxNewRepository) GetDashboardSummary(startDate, endDate time.Time) (*domain.DashboardSummaryResponse, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DashboardSummaryResponse), args.Error(1)
}

func (m *MockTaxNewRepository) UpsertBusiness(business *domain.TaxBusiness) error {
	args := m.Called(business)
	return args.Error(0)
}

func (m *MockTaxNewRepository) GetUserInformationByPhoneOrEmail(phone *string, email string) (*domain.UserInformation, error) {
	args := m.Called(phone, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserInformation), args.Error(1)
}

func TestUploadKTBFile(t *testing.T) {
	repo := new(MockTaxNewRepository)
	mailSender := new(MockEmailSender)
	uc := NewTaxNewUseCase(repo, mailSender, "099400016485800")

	adminID := uuid.New()
	filename := "ktb_reconciliation_sample.txt"

	fileContent := []byte(`010205561001234  CHONBURI PAO                            202605220062286032092                                          
0220260522083015123456701           20260500            0000000075000                                                   
0220260522091530765432102           20260200            0000000408600                                                   
09000002000000000483600                                                                                                 
`)

	mockBusiness := &domain.TaxBusiness{
		ID:                uuid.New(),
		BusinessRegNumber: "1234567",
		NameTH:            "บริษัท ทดสอบ จำกัด",
		TaxType:           "hotel_fee",
	}

	mockDecl1 := &domain.TaxDeclaration{
		ID:                 uuid.New(),
		BusinessID:         mockBusiness.ID,
		BusinessRegNumber:  "1234567",
		TaxType:            "hotel_fee",
		TaxMonth:           5,
		TaxYear:            2026,
		DeclarationVersion: 1,
		CalculatedTax:      750.00,
		PayerEmail:         "payer1@example.com",
		Ref1:               "123456701",
		Ref2:               "20260500",
		Business:           mockBusiness,
	}

	repo.On("GetDeclarationByRefs", "123456701", "20260500").Return(mockDecl1, nil)
	repo.On("GetDeclarationByRefs", "765432102", "20260200").Return(nil, errors.New("not found"))
	repo.On("GetUserInformationByPhoneOrEmail", mock.Anything, mock.Anything).Return(nil, nil)

	repo.On("CreateReconciliationBatch", mock.Anything).Return(nil)
	repo.On("CreateReconciliationRecord", mock.Anything).Return(nil)
	repo.On("UpdateDeclaration", mock.Anything).Return(nil)

	mailSender.On("SendHTML", []string{"payer1@example.com"}, mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.UploadKTBFile(filename, fileContent, adminID)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, 2, resp.TotalRecords)
	assert.Equal(t, 1, resp.MatchedRecords)
	assert.Equal(t, 1, resp.UnmatchedRecords)
	assert.Equal(t, 4836.00, resp.TotalAmount)

	repo.AssertExpectations(t)
	mailSender.AssertExpectations(t)
}

func TestUploadKTBFileSmartFallback(t *testing.T) {
	repo := new(MockTaxNewRepository)
	mailSender := new(MockEmailSender)
	uc := NewTaxNewUseCase(repo, mailSender, "099400016485800")

	adminID := uuid.New()
	filename := "ktb_combined_test_reconciliation.txt"

	fileContent := []byte(`010205561001234 CHONBURI PAO 202605280062286032092
0220260528120928123456701 20260501 0000000030000
0220260528121034765432102 20260501 0000000227000
090000020000000257000
`)

	mockBusiness1 := &domain.TaxBusiness{
		ID:                uuid.New(),
		BusinessRegNumber: "1234567",
		NameTH:            "ร้านค้าโรงแรมทดสอบ",
		TaxType:           "hotel_fee",
	}

	mockDecl1 := &domain.TaxDeclaration{
		ID:                 uuid.New(),
		BusinessID:         mockBusiness1.ID,
		BusinessRegNumber:  "1234567",
		TaxType:            "hotel_fee",
		TaxMonth:           5,
		TaxYear:            2026,
		DeclarationVersion: 1,
		CalculatedTax:      300.00,
		PayerEmail:         "hotel@example.com",
		Ref1:               "123456701",
		Ref2:               "20260501",
		Business:           mockBusiness1,
	}

	mockBusiness2 := &domain.TaxBusiness{
		ID:                uuid.New(),
		BusinessRegNumber: "7654321",
		NameTH:            "ร้านค้าปั๊มน้ำมันทดสอบ",
		TaxType:           "oil_gas_tax",
	}

	mockDecl2 := &domain.TaxDeclaration{
		ID:                 uuid.New(),
		BusinessID:         mockBusiness2.ID,
		BusinessRegNumber:  "7654321",
		TaxType:            "oil_gas_tax",
		TaxMonth:           5,
		TaxYear:            2026,
		DeclarationVersion: 1,
		CalculatedTax:      2270.00,
		PayerEmail:         "oil@example.com",
		Ref1:               "765432102",
		Ref2:               "20260501",
		Business:           mockBusiness2,
	}

	repo.On("GetDeclarationByRefs", "123456701", "20260501").Return(mockDecl1, nil)
	repo.On("GetDeclarationByRefs", "765432102", "20260501").Return(mockDecl2, nil)
	repo.On("GetUserInformationByPhoneOrEmail", mock.Anything, mock.Anything).Return(nil, nil)

	repo.On("CreateReconciliationBatch", mock.Anything).Return(nil)
	repo.On("CreateReconciliationRecord", mock.Anything).Return(nil)
	repo.On("UpdateDeclaration", mock.Anything).Return(nil)

	mailSender.On("SendHTML", []string{"hotel@example.com"}, mock.Anything, mock.Anything).Return(nil)
	mailSender.On("SendHTML", []string{"oil@example.com"}, mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.UploadKTBFile(filename, fileContent, adminID)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, 2, resp.TotalRecords)
	assert.Equal(t, 2, resp.MatchedRecords)
	assert.Equal(t, 0, resp.UnmatchedRecords)
	assert.Equal(t, 2570.00, resp.TotalAmount)

	repo.AssertExpectations(t)
	mailSender.AssertExpectations(t)
}

func TestUploadElaasFile(t *testing.T) {
	repo := new(MockTaxNewRepository)
	mailSender := new(MockEmailSender)
	uc := NewTaxNewUseCase(repo, mailSender, "099400016485800")

	adminID := uuid.New()
	filename := "elaas_daily_summary_sample.csv"
	fileContent := []byte(`summary_date,tax_type,total_amount,transaction_count
2026-05-22,hotel_fee,45000.00,22
2026-05-22,oil_gas_tax,89000.00,14
`)

	repo.On("UpsertElaasSummary", mock.Anything).Return(nil)

	count, err := uc.UploadElaasFile(filename, fileContent, adminID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	repo.AssertExpectations(t)
}
