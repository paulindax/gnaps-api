package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
)

type FinanceAccountService struct {
	accountRepo *repositories.FinanceAccountRepository
}

func NewFinanceAccountService(accountRepo *repositories.FinanceAccountRepository) *FinanceAccountService {
	return &FinanceAccountService{accountRepo: accountRepo}
}

func (s *FinanceAccountService) GetAccountByID(id uint) (*models.FinanceAccount, error) {
	account, err := s.accountRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("finance account not found")
	}
	return account, nil
}

func (s *FinanceAccountService) ListAccounts(filters map[string]interface{}, page, limit int) ([]models.FinanceAccount, int64, error) {
	return s.accountRepo.List(filters, page, limit)
}

func (s *FinanceAccountService) CreateAccount(account *models.FinanceAccount) error {
	// Validate required fields
	if account.Name == nil || *account.Name == "" {
		return errors.New("name is required")
	}
	if account.Code == nil || *account.Code == "" {
		return errors.New("code is required")
	}

	// Check if code already exists
	exists, err := s.accountRepo.CodeExists(*account.Code, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("finance account with this code already exists")
	}

	// Set defaults
	account.IsDeleted = false

	return s.accountRepo.Create(account)
}

func (s *FinanceAccountService) UpdateAccount(id uint, updates map[string]interface{}) error {
	// Verify account exists
	account, err := s.accountRepo.FindByID(id)
	if err != nil {
		return errors.New("finance account not found")
	}

	// Check if code is being changed and if new code already exists
	if code, ok := updates["code"]; ok {
		codeStr := code.(string)
		if codeStr != "" && (account.Code == nil || codeStr != *account.Code) {
			exists, err := s.accountRepo.CodeExists(codeStr, &id)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("finance account with this code already exists")
			}
		}
	}

	return s.accountRepo.Update(id, updates)
}

func (s *FinanceAccountService) DeleteAccount(id uint) error {
	_, err := s.accountRepo.FindByID(id)
	if err != nil {
		return errors.New("finance account not found")
	}

	return s.accountRepo.Delete(id)
}

// ============================================
// Owner-based methods for data filtering
// ============================================

func (s *FinanceAccountService) GetAccountByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.FinanceAccount, error) {
	return s.accountRepo.FindByIDWithOwner(id, ownerCtx)
}

func (s *FinanceAccountService) ListAccountsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.FinanceAccount, int64, error) {
	return s.accountRepo.ListWithOwner(filters, page, limit, ownerCtx)
}

func (s *FinanceAccountService) CreateAccountWithOwner(account *models.FinanceAccount, ownerCtx *utils.OwnerContext) error {
	// Validate required fields
	if account.Name == nil || *account.Name == "" {
		return errors.New("name is required")
	}
	if account.Code == nil || *account.Code == "" {
		return errors.New("code is required")
	}

	// Check if code already exists
	exists, err := s.accountRepo.CodeExists(*account.Code, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("finance account with this code already exists")
	}

	// Set defaults
	account.IsDeleted = false

	return s.accountRepo.CreateWithOwner(account, ownerCtx)
}

func (s *FinanceAccountService) UpdateAccountWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	// Check if code is being changed and if new code already exists
	if code, ok := updates["code"]; ok {
		codeStr := code.(string)
		if codeStr != "" {
			exists, err := s.accountRepo.CodeExists(codeStr, &id)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("finance account with this code already exists")
			}
		}
	}

	return s.accountRepo.UpdateWithOwner(id, updates, ownerCtx)
}

func (s *FinanceAccountService) DeleteAccountWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	return s.accountRepo.DeleteWithOwner(id, ownerCtx)
}
