package bankaccounts

import (
	"context"

	pbBankBranches "davensi.com/core/gen/bankbranches"
	pbUoms "davensi.com/core/gen/uoms"
	"github.com/bufbuild/connect-go"
)

type BankAccountRelationships struct {
	BankBranch *pbBankBranches.BankBranch
	Currency   *pbUoms.UoM
}

// For BankBranch
func (s *ServiceServer) GetRelationship(
	selectBankBranch *pbBankBranches.Select,
	selectCurrency *pbUoms.Select,
) BankAccountRelationships {
	bankBranchChan := make(chan *pbBankBranches.BankBranch)
	currencyChan := make(chan *pbUoms.UoM)

	// BankBranch field
	go func() {
		var existBankBranch *pbBankBranches.BankBranch
		if selectBankBranch != nil {
			getBankBranchResponse, err := s.bankBranchesSS.Get(context.Background(), &connect.Request[pbBankBranches.GetRequest]{
				Msg: &pbBankBranches.GetRequest{
					Select: selectBankBranch,
				},
			})
			if err == nil {
				existBankBranch = getBankBranchResponse.Msg.GetBankbranch()
			}
		}
		bankBranchChan <- existBankBranch
	}()

	// Currency field
	go func() {
		var existCurrency *pbUoms.UoM
		if selectCurrency != nil {
			getCurrencyResponse, err := s.uomsSS.Get(context.Background(), &connect.Request[pbUoms.GetRequest]{
				Msg: &pbUoms.GetRequest{
					Select: selectCurrency,
				},
			})
			if err == nil {
				existCurrency = getCurrencyResponse.Msg.GetUom()
			}
		}
		currencyChan <- existCurrency
	}()

	return BankAccountRelationships{
		BankBranch: <-bankBranchChan,
		Currency:   <-currencyChan,
	}
}
