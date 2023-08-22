package countries

import (
	"context"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbUoMs "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
)

func ValidateSelect(msg *pbCountries.Select, method string) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_code must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbCountries.Select_ByCode:
		if msg.GetByCode() == "" {
			return errUpdate.UpdateMessage("by_code must be specified")
		}
	case *pbCountries.Select_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) IsCountryUniq(countryTypeSymbol string) (isUniq bool, errCode pbCommon.ErrorCode) {
	country, err := s.Get(context.Background(), &connect.Request[pbCountries.GetRequest]{
		Msg: &pbCountries.GetRequest{
			Select: &pbCountries.Select{
				Select: &pbCountries.Select_ByCode{
					ByCode: countryTypeSymbol,
				},
			},
		},
	})

	if err == nil || country.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
	}

	if country.Msg.GetError() != nil && country.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, country.Msg.GetError().Code
	}

	return true, country.Msg.GetError().Code
}

// for Create gRPC
func (s *ServiceServer) ValidateCreate(msg *pbCountries.CreateRequest) *common.ErrWithCode {
	// Verify that Type and Symbol are specified
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if msg.Code == "" {
		return errValidate.UpdateMessage("code must be specified")
	}

	if isUniq, errCode := s.IsCountryUniq(msg.GetCode()); !isUniq {
		return errValidate.UpdateCode(errCode).UpdateMessage("code have been used")
	}

	return nil
}

func (s *ServiceServer) ValidateUpdateValue(
	oldCountry *pbCountries.Country,
	msg *pbCountries.UpdateRequest,
) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	if msg.Code != nil && msg.GetCode() != oldCountry.Code {
		if isUniq, errCode := s.IsCountryUniq(msg.GetCode()); !isUniq {
			return errValidate.UpdateCode(errCode).UpdateMessage("code have been used")
		}
	}

	return nil
}

func validateUpsertCountriesUoms(country *pbCountries.Country, selectUoMs *pbUoMs.SelectList) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"upsert",
		"set uoms",
		"",
	)

	if country == nil || country.GetId() == "" {
		return errValidate.UpdateMessage("country must be specified")
	}

	if len(selectUoMs.GetList()) == 0 {
		return errValidate.UpdateMessage("list must be specified")
	}

	for index, selectUom := range selectUoMs.GetList() {
		if selectUom.GetById() == "" {
			return errValidate.UpdateMessage(
				fmt.Sprintf("select index: %d have error: 'cannot get select uom by id'", index),
			)
		}
	}

	return nil
}

func (s *ServiceServer) validateUpsertCountriesRelationship(
	selectCryptos *pbUoMs.SelectList,
	selectFiats *pbUoMs.SelectList,
) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"upsert",
		"set uoms",
		"",
	)

	relationship := s.GetRelationship(selectCryptos, selectFiats)

	if selectCryptos != nil && len(selectCryptos.GetList()) > 0 {
		if len(relationship.Cryptos) == 0 {
			errValidate.UpdateMessage("cryptos not found")
		} else {
			cryptosByID := &pbUoMs.SelectList{}
			for _, crypto := range relationship.Cryptos {
				cryptosByID.List = append(cryptosByID.List, &pbUoMs.Select{
					Select: &pbUoMs.Select_ById{
						ById: crypto.Uom.Id,
					},
				})
			}
		}
	}

	if selectFiats != nil && len(selectFiats.GetList()) > 0 {
		if len(relationship.Fiats) == 0 {
			errValidate.UpdateMessage("fiats not found")
		} else {
			fiatsByID := &pbUoMs.SelectList{}
			for _, crypto := range relationship.Fiats {
				fiatsByID.List = append(fiatsByID.List, &pbUoMs.Select{
					Select: &pbUoMs.Select_ById{
						ById: crypto.Uom.Id,
					},
				})
			}
		}
	}

	return errValidate
}
