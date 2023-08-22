package cryptocategories

import (
	"context"
	"errors"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbCryptocategories "davensi.com/core/gen/cryptocategories"
	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsCryptoCategoryUniq(cryptoCategoryTypeSymbol string) (isUniq bool, errno pbCommon.ErrorCode) {
	cryptoCategory, err := s.Get(context.Background(), &connect.Request[pbCryptocategories.GetRequest]{
		Msg: &pbCryptocategories.GetRequest{
			Select: &pbCryptocategories.GetRequest_ByName{
				ByName: cryptoCategoryTypeSymbol,
			},
		},
	})

	if err == nil || cryptoCategory.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
	}

	if cryptoCategory.Msg.GetError() != nil && cryptoCategory.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, cryptoCategory.Msg.GetError().Code
	}

	return true, cryptoCategory.Msg.GetError().Code
}

// for Create gRPC
func (s *ServiceServer) ValidateCreate(msg *pbCryptocategories.CreateRequest) (errno pbCommon.ErrorCode, err error) {
	// Verify that Type and Symbol are specified
	if msg.Name == "" {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("creating '%s' type and symbol must be specified", _entityName)
	}

	if isUniq, errno := s.IsCryptoCategoryUniq(msg.GetName()); !isUniq {
		return errno, fmt.Errorf("create %s code = '%s' Code", _entityName, msg.GetName())
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

// for Update gRPC
func validateQueryUpdate(msg *pbCryptocategories.UpdateRequest) error {
	switch msg.GetSelect().(type) {
	case *pbCryptocategories.UpdateRequest_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errors.New("id must be specified")
		}
	case *pbCryptocategories.UpdateRequest_ByName:
		// Verify that Type and Symbol are specified
		if msg.GetByName() == "" {
			return errors.New("type and symbol must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) ValidateUpdateValue(
	oldCryptoCategory *pbCryptocategories.CryptoCategory,
	msg *pbCryptocategories.UpdateRequest,
) (pbCommon.ErrorCode, error) {
	checkName := oldCryptoCategory.Name
	isUpdateName := false

	if msg.Name != nil {
		checkName = msg.GetName()
		isUpdateName = true
	}

	if isUpdateName {
		checkCryptoCategory, err := s.Get(context.Background(), &connect.Request[pbCryptocategories.GetRequest]{
			Msg: &pbCryptocategories.GetRequest{
				Select: &pbCryptocategories.GetRequest_ByName{
					ByName: checkName,
				},
			},
		})
		if err == nil || checkCryptoCategory.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
			return pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY, err
		}
		// Verify that Type/Symbol does not exist (actually expecting ERROR_CODE_NOT_FOUND)
		if checkCryptoCategory.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
			return checkCryptoCategory.Msg.GetError().Code, err
		}
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}
