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

// MockEmailSender is a mock email sender
type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendHTML(to []string, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

// MockTaxNewRepository is a mock repository implementation
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

func (m *MockTaxNewRepository) ListDeclarations(taxType, status, search string, limit, offset int) ([]domain.TaxDeclaration, int64, error) {
	args := m.Called(taxType, status, search, limit, offset)
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

func TestUploadKTBFile(t *testing.T) {
	repo := new(MockTaxNewRepository)
	mailSender := new(MockEmailSender)
	uc := NewTaxNewUseCase(repo, mailSender, "099400016485800")

	adminID := uuid.New()
	filename := "ktb_reconciliation_sample.txt"
	
	// Sample content similar to what we observed:
	// Line 1: Header (record type 01)
	// Line 2: Detail record (record type 02), match with ref1=123456701, ref2=20260500, amount=750.00
	// Line 3: Detail record (record type 02), mismatch or match with ref1=765432102, ref2=20260200, amount=4086.00
	// Line 4: Footer (record type 09)
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

	// Set expectations
	// First line is detail record for ref1: 123456701, ref2: 20260500, amount: 750.00
	repo.On("GetDeclarationByRefs", "123456701", "20260500").Return(mockDecl1, nil)
	// Second line is detail record for ref1: 765432102, ref2: 20260200, amount: 4086.00 (fails match or not found)
	repo.On("GetDeclarationByRefs", "765432102", "20260200").Return(nil, errors.New("not found"))

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
