package banks

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"

	pbBanks "davensi.com/core/gen/banks"
	pbCommon "davensi.com/core/gen/common"

	"davensi.com/core/internal/common"
)

type BankRelationshipIds struct {
	BankID *uuid.UUID
}

func (s *ServiceServer) GetBankRelationshipIds(
	selectBank *pbBanks.Select,
) BankRelationshipIds {
	bankIDChan := make(chan *uuid.UUID)

	go func() {
		var existBankID *uuid.UUID
		if selectBank != nil {
			existBank, _ := s.getBankSelect(selectBank)
			if existBank != nil {
				parse, err := uuid.Parse(existBank.GetId())
				if err != nil {
					existBankID = &parse
				}
			}
		}
		bankIDChan <- existBankID
	}()

	return BankRelationshipIds{
		BankID: <-bankIDChan,
	}
}

func (s *ServiceServer) getBankSelect(req *pbBanks.Select) (*pbBanks.Bank, *common.ErrWithCode) {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"getting",
		_entityName,
		"",
	)

	bankInput := &pbBanks.GetRequest{
		Select: req,
	}

	type magicType string
	var magicKey magicType = "magicValue"
	bankResponse, errBank := GetSingletonServiceServer(s.db).Get(
		context.WithValue(context.Background(), magicKey, magicKey),
		connect.NewRequest(bankInput),
	)

	if errBank != nil {
		return nil, errGet.UpdateMessage(errBank.Error())
	}

	return bankResponse.Msg.GetBank(), nil
}
