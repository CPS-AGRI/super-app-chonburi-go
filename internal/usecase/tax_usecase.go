package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"super-app-chonburi-go/internal/domain"
	"time"

	"github.com/google/uuid"
)

type taxUseCase struct {
	taxRepo domain.TaxRepository
	// We might need user repo here too, but for now we'll assume taxRepo can find users or we use raw DB
}

func NewTaxUseCase(taxRepo domain.TaxRepository) domain.TaxUseCase {
	return &taxUseCase{taxRepo: taxRepo}
}

func (u *taxUseCase) ImportTaxRecords(importData *domain.ModuleOnlineTaxPayment, records []domain.ModuleOnlineTaxPaymentInformation, adminID uuid.UUID) (int, int, []string) {
	var successCount, errorCount int
	var errorMessages []string

	// 1. Create Import Head
	importData.ID = uuid.New()
	importData.AdminUserId = adminID
	importData.CreatedDate = time.Now()
	importData.UpdatedDate = time.Now()

	if err := u.taxRepo.CreateImport(importData); err != nil {
		return 0, len(records), []string{fmt.Sprintf("Failed to create import head: %v", err)}
	}

	// 2. Process Records in Background (Goroutine)
	// For now, we'll run it synchronously but designed to be async if needed
	// The caller should ideally wrap this in a goroutine if they want immediate response
	// But according to requirements, we should use Goroutines here.
	
	go func() {
		for _, record := range records {
			err := u.processSingleRecord(importData, record, adminID)
			if err != nil {
				errorCount++
				errorMessages = append(errorMessages, fmt.Sprintf("Record %s: %v", record.Name, err))
			} else {
				successCount++
			}
		}
		log.Printf("Import %s completed: %d success, %d errors", importData.Name, successCount, errorCount)
		// TODO: Save import summary/report somewhere so admin can view it later
	}()

	return 0, 0, nil // Return immediately as it's background job
}

func (u *taxUseCase) processSingleRecord(importData *domain.ModuleOnlineTaxPayment, record domain.ModuleOnlineTaxPaymentInformation, adminID uuid.UUID) error {
	// 1. Check for duplicates (Ref1 & Ref2)
	existing, err := u.taxRepo.GetInformationByRefs(record.ReferenceNumber1, record.ReferenceNumber2)
	if err != nil {
		return err
	}

	if existing != nil {
		// Rule: If already paid, REJECT update
		if existing.Status == domain.TaxStatusCompleted {
			return fmt.Errorf("duplicate reference and already paid")
		}
		// Rule: If pending, OVERWRITE
		oldData, _ := json.Marshal(existing)
		
		existing.Name = record.Name
		existing.Amount = record.Amount
		existing.PaymentDueDate = record.PaymentDueDate
		existing.IdentityNumber = record.IdentityNumber
		existing.UpdatedDate = time.Now()
		
		// Re-attempt mapping if identity number changed
		u.attemptAutoMapping(existing)

		if err := u.taxRepo.UpdateInformation(existing); err != nil {
			return err
		}

		newData, _ := json.Marshal(existing)
		u.taxRepo.CreateLog(&domain.ModuleOnlineTaxPaymentLog{
			ID:                                  uuid.New(),
			AdminUserId:                         adminID,
			ModuleOnlineTaxPaymentInformationId: existing.ID,
			Action:                              "IMPORT_OVERWRITE",
			OldData:                             oldData,
			NewData:                             newData,
			CreatedDate:                         time.Now(),
		})
		return nil
	}

	// 2. New Record
	record.ID = uuid.New()
	record.ModuleOnlineTaxPaymentId = importData.ID
	record.Status = domain.TaxStatusPending
	record.CreatedDate = time.Now()
	record.UpdatedDate = time.Now()
	
	// Set initial DocumentID (logic can be more complex)
	record.DocumentId = fmt.Sprintf("TAX-%s-%d", importData.Year, time.Now().UnixNano()%100000)

	// 3. Attempt Auto-Mapping with IdentityNumber
	u.attemptAutoMapping(&record)

	if err := u.taxRepo.CreateInformation(&record); err != nil {
		return err
	}

	newData, _ := json.Marshal(record)
	u.taxRepo.CreateLog(&domain.ModuleOnlineTaxPaymentLog{
		ID:                                  uuid.New(),
		AdminUserId:                         adminID,
		ModuleOnlineTaxPaymentInformationId: record.ID,
		Action:                              "IMPORT_CREATE",
		OldData:                             json.RawMessage("null"),
		NewData:                             newData,
		CreatedDate:                         time.Now(),
	})

	return nil
}

func (u *taxUseCase) attemptAutoMapping(record *domain.ModuleOnlineTaxPaymentInformation) {
	// TODO: Find user by IdentityNumber
	// For now, we'll set LinkStatus to manual/not_linked if not found
	// In a real implementation, we would query the users table
	record.LinkStatus = domain.TaxLinkStatusNotLinked
	
	// Logic to find user...
	// if found {
	//    record.UserId = &user.ID
	//    record.LinkStatus = domain.TaxLinkStatusAutoLinked
	// }
}

func (u *taxUseCase) GetMyTaxes(identityNumber string) ([]domain.ModuleOnlineTaxPaymentInformation, error) {
	res, err := u.taxRepo.GetInformationsPaginated(domain.TaxQuery{
		IdentityNumber: identityNumber,
		PageNumber:     1,
		PageSize:       100, // Show all for mobile list
	})
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (u *taxUseCase) UpdateTaxStatus(id uuid.UUID, status string, adminID uuid.UUID, receiptUrl *string) error {
	existing, err := u.taxRepo.GetInformationByID(id.String())
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("tax record not found")
	}

	oldData, _ := json.Marshal(existing)

	existing.Status = status
	if receiptUrl != nil {
		existing.ReceiptUrl = receiptUrl
	}
	existing.UpdatedDate = time.Now()

	if err := u.taxRepo.UpdateInformation(existing); err != nil {
		return err
	}

	newData, _ := json.Marshal(existing)
	u.taxRepo.CreateLog(&domain.ModuleOnlineTaxPaymentLog{
		ID:                                  uuid.New(),
		AdminUserId:                         adminID,
		ModuleOnlineTaxPaymentInformationId: existing.ID,
		Action:                              "UPDATE_STATUS",
		OldData:                             oldData,
		NewData:                             newData,
		CreatedDate:                         time.Now(),
	})

	return nil
}

func (u *taxUseCase) LinkUser(infoID uuid.UUID, userID uuid.UUID, adminID uuid.UUID) error {
	existing, err := u.taxRepo.GetInformationByID(infoID.String())
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("tax record not found")
	}

	oldData, _ := json.Marshal(existing)

	existing.UserId = &userID
	existing.LinkStatus = domain.TaxLinkStatusManualLinked
	existing.UpdatedDate = time.Now()

	if err := u.taxRepo.UpdateInformation(existing); err != nil {
		return err
	}

	newData, _ := json.Marshal(existing)
	u.taxRepo.CreateLog(&domain.ModuleOnlineTaxPaymentLog{
		ID:                                  uuid.New(),
		AdminUserId:                         adminID,
		ModuleOnlineTaxPaymentInformationId: existing.ID,
		Action:                              "LINK_USER",
		OldData:                             oldData,
		NewData:                             newData,
		CreatedDate:                         time.Now(),
	})

	return nil
}
