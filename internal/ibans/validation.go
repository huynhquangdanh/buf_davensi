package ibans

import (
	"context"
	"errors"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbIbans "davensi.com/core/gen/ibans"
	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *ServiceServer) IsIbanUniq(
	country *pbCountries.Select,
	validity *timestamppb.Timestamp,
) (isUniq bool, errno pbCommon.ErrorCode) {
	iban, err := s.Get(context.Background(), &connect.Request[pbIbans.GetRequest]{
		Msg: &pbIbans.GetRequest{
			Select: &pbIbans.Select{
				Select: &pbIbans.Select_ByCountryValidity{
					ByCountryValidity: &pbIbans.CountryValidity{
						Country:   country,
						ValidFrom: validity,
					},
				},
			},
		},
	})

	if err == nil || iban.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if iban.Msg.GetError() != nil && iban.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, iban.Msg.GetError().Code
	}

	return true, iban.Msg.GetError().Code
}

// for Create gRPC
func (s *ServiceServer) validateCreation(msg *pbIbans.CreateRequest) (errno pbCommon.ErrorCode, err error) {
	if msg.Country == nil {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("creating '%s' country must be specified", _entityName)
	}

	if msg.Algorithm == pbIbans.Algorithm_ALGORITHM_UNSPECIFIED {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("creating '%s' algorithm must be specified", _entityName)
	}

	if msg.Format == "" {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("creating '%s' format must be specified", _entityName)
	}

	ibanRL := s.GetRelationship(msg.GetCountry())
	if ibanRL.country == nil {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("country does not exist")
	}

	msg.Country = &pbCountries.Select{
		Select: &pbCountries.Select_ById{
			ById: ibanRL.country.Id,
		},
	}

	if msg.Validity == nil {
		msg.Validity = timestamppb.Now()
	}

	if isUniq, errno := s.IsIbanUniq(msg.Country, msg.Validity); !isUniq {
		return errno, fmt.Errorf("create %s country = '%v' validity = '%v'", _entityName, msg.GetCountry(), msg.GetValidity())
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

// for Update gRPC
func validateQueryUpdate(msg *pbIbans.UpdateRequest) error {
	switch msg.GetSelect().(type) {
	case *pbIbans.UpdateRequest_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errors.New("id must be specified")
		}
	case *pbIbans.UpdateRequest_ByCountryValidity:
		// Verify that Type and Symbol are specified
		if msg.GetByCountryValidity() == nil {
			return errors.New("country validity must be specified")
		}

		if msg.GetByCountryValidity().GetCountry() == nil {
			return errors.New("country must be specified")
		}
		if msg.GetByCountryValidity().GetValidFrom() == nil {
			return errors.New("valid_from must be specified")
		}

		if msg.GetByCountryValidity().GetCountry().GetByCode() == "" ||
			msg.GetByCountryValidity().GetCountry().GetById() == "" {
			return errors.New("country select(id or code) must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateMsgUpdate(
	oldIban *pbIbans.IBAN,
	msg *pbIbans.UpdateRequest,
) (pbCommon.ErrorCode, error) {
	checkCountry := &pbCountries.Select{
		Select: &pbCountries.Select_ById{
			ById: oldIban.GetCountry().Id,
		},
	}
	checkValidity := oldIban.ValidFrom
	isCheckUniq := false

	if msg.Country != nil {
		checkCountry = msg.GetCountry()
		isCheckUniq = true
	}

	if msg.Validity != nil {
		checkValidity = msg.GetValidity()
		isCheckUniq = true
	}

	if isCheckUniq {
		isIbanUniq, errno := s.IsIbanUniq(checkCountry, checkValidity)
		if !isIbanUniq {
			return errno, errors.New("country and validity have been used")
		}
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

func ValidateSelect(msg *pbIbans.Select, method string) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)
	if msg == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbIbans.Select_ByCountryValidity:
		if msg.GetByCountryValidity() == nil {
			return errUpdate.UpdateMessage("by_type_name must be specified")
		}
		if msg.GetByCountryValidity().Country == nil || msg.GetByCountryValidity().ValidFrom == nil {
			return errUpdate.UpdateMessage("type and name must be specified")
		}
	case *pbIbans.Select_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}
