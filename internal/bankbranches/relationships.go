package bankbranches

import (
	"context"

	pbBanks "davensi.com/core/gen/banks"
	"github.com/bufbuild/connect-go"
)

type BankBranchRelationships struct {
	Bank *pbBanks.Bank
}

// For Incorporation Country, Currency1, Currency2, Currency3
func (s *ServiceServer) GetRelationship(
	selectBank *pbBanks.Select,
) BankBranchRelationships {
	bankChan := make(chan *pbBanks.Bank)

	// Bank field
	go func() {
		var existBank *pbBanks.Bank
		if selectBank != nil {
			getBankBranchResponse, err := s.banksSS.Get(context.Background(), &connect.Request[pbBanks.GetRequest]{
				Msg: &pbBanks.GetRequest{
					Select: selectBank,
				},
			})
			if err == nil {
				existBank = getBankBranchResponse.Msg.GetBank()
			}
		}
		bankChan <- existBank
	}()

	return BankBranchRelationships{
		Bank: <-bankChan,
	}
}
