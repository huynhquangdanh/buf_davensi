package legalentities

import (
	"context"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbUoms "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/countries"
	"davensi.com/core/internal/uoms"

	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsLegalEntityUniq(selectByName *pbLegalEntities.Select_ByName) (isUniq bool, errno pbCommon.ErrorCode) {
	legalEntity, err := s.Get(context.Background(), &connect.Request[pbLegalEntities.GetRequest]{
		Msg: &pbLegalEntities.GetRequest{
			Select: &pbLegalEntities.Select{
				Select: selectByName,
			},
		},
	})

	if err == nil || legalEntity.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if legalEntity.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, legalEntity.Msg.GetError().Code
	}

	return true, legalEntity.Msg.GetError().Code
}

func (s *ServiceServer) ValidateAddAddresses(req *pbLegalEntities.AddAddressesRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "user.User must be specified")
	}
	return
}

func (s *ServiceServer) ValidateSetAddresses(req *pbLegalEntities.SetAddressesRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "Legal Entity must be specified")
	}
	return
}

func (s *ServiceServer) ValidateUpdateAddress(req *pbLegalEntities.UpdateAddressRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "Legal Entity must be specified")
	}
	return
}

func (s *ServiceServer) ValidateRemoveAddresses(req *pbLegalEntities.RemoveAddressesRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "Legal Entity must be specified")
	}
	return
}

func (s *ServiceServer) ValidateSetContacts(req *pbLegalEntities.SetContactsRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "Legal Entity must be specified")
	}
	return
}

func (s *ServiceServer) ValidateAddContacts(req *pbLegalEntities.AddContactsRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "user.User must be specified")
	}
	return
}

func (s *ServiceServer) ValidateUpdateContact(req *pbLegalEntities.UpdateContactRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "Legal Entity must be specified")
	}
	return
}

func (s *ServiceServer) ValidateRemoveContacts(req *pbLegalEntities.RemoveContactsRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.LegalEntity.GetById() == "" && req.LegalEntity.GetByName() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "Legal Entity must be specified")
	}
	return
}

func ValidateSelect(selectLegalEntity *pbLegalEntities.Select, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_entityName,
		"",
	)

	if selectLegalEntity == nil {
		return errValidate.UpdateMessage("Select by ID or Name must be specified")
	}

	if selectLegalEntity.Select == nil {
		return errValidate.UpdateMessage("Select by ID or Name must be specified")
	}

	switch selectLegalEntity.GetSelect().(type) {
	case *pbLegalEntities.Select_ById:
		if selectLegalEntity.GetById() == "" {
			return errValidate.UpdateMessage("by_id must be specified")
		}
	case *pbLegalEntities.Select_ByName:
		if selectLegalEntity.GetByName() == "" {
			return errValidate.UpdateMessage("by_name must be specified")
		}
	}

	return nil
}

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbLegalEntities.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	// Verify that Name, Type, Incorporation Country, Currency1 are specified
	if msg.Name == "" {
		return errCreation.UpdateMessage("name must be specified")
	}

	if msg.Type == pbLegalEntities.Type_TYPE_UNSPECIFIED {
		return errCreation.UpdateMessage("type must be specified")
	}

	if errSelectSource := countries.ValidateSelect(msg.GetIncorporationCountry(), "creating"); errSelectSource != nil {
		return errSelectSource
	}

	if errSelectSource := uoms.ValidateSelect(msg.GetCurrency1(), "creating"); errSelectSource != nil {
		return errSelectSource
	}

	legalEntityRl := s.GetRelationship(
		msg.GetIncorporationCountry(),
		msg.GetCurrency1(),
		msg.GetCurrency2(),
		msg.GetCurrency3(),
	)

	if legalEntityRl.Country == nil {
		return errCreation.UpdateMessage("country does not exist")
	}

	if legalEntityRl.UoM1 == nil {
		return errCreation.UpdateMessage("currency1 (uom1) does not exist")
	}

	msg.IncorporationCountry = &pbCountries.Select{
		Select: &pbCountries.Select_ById{
			ById: legalEntityRl.Country.Id,
		},
	}

	msg.Currency1 = &pbUoms.Select{
		Select: &pbUoms.Select_ById{
			ById: legalEntityRl.UoM1.Id,
		},
	}

	if msg.Currency2 != nil && legalEntityRl.UoM2 == nil {
		return errCreation.UpdateMessage("currency2 (uom2) does not exist")
	} else if legalEntityRl.UoM2 != nil {
		msg.Currency2 = &pbUoms.Select{
			Select: &pbUoms.Select_ById{
				ById: legalEntityRl.UoM2.Id,
			},
		}
	}

	if msg.Currency3 != nil && legalEntityRl.UoM3 == nil {
		return errCreation.UpdateMessage("currency3 (uom3) does not exist")
	} else if legalEntityRl.UoM3 != nil {
		msg.Currency3 = &pbUoms.Select{
			Select: &pbUoms.Select_ById{
				ById: legalEntityRl.UoM3.Id,
			},
		}
	}

	if isUniq, errCode := s.IsLegalEntityUniq(&pbLegalEntities.Select_ByName{
		ByName: msg.GetName(),
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("name has been used")
	}

	return nil
}

// for Update gRPC
// Check for WHERE clause, whether the relationships exist
func (s *ServiceServer) validateUpdateQuery(msg *pbLegalEntities.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	if errValidateSelect := ValidateSelect(msg.GetSelect(), "updating"); errValidateSelect != nil {
		return errValidateSelect
	}

	legalEntityRl := s.GetRelationship(
		msg.GetIncorporationCountry(),
		msg.GetCurrency1(),
		msg.GetCurrency2(),
		msg.GetCurrency3(),
	)

	if msg.IncorporationCountry != nil && legalEntityRl.Country == nil {
		return errUpdate.UpdateMessage("incorporation country does not exist")
	}

	if msg.Currency1 != nil && legalEntityRl.UoM1 == nil {
		return errUpdate.UpdateMessage("currency1 does not exist")
	}

	if msg.Currency2 != nil && legalEntityRl.UoM2 == nil {
		return errUpdate.UpdateMessage("currency2 does not exist")
	}

	if msg.Currency3 != nil && legalEntityRl.UoM3 == nil {
		return errUpdate.UpdateMessage("currency3 does not exist")
	}

	if legalEntityRl.Country != nil {
		msg.IncorporationCountry = &pbCountries.Select{
			Select: &pbCountries.Select_ById{
				ById: legalEntityRl.Country.Id,
			},
		}
	}

	if legalEntityRl.UoM1 != nil {
		msg.Currency1 = &pbUoms.Select{
			Select: &pbUoms.Select_ById{
				ById: legalEntityRl.UoM1.Id,
			},
		}
	}

	if legalEntityRl.UoM2 != nil {
		msg.Currency2 = &pbUoms.Select{
			Select: &pbUoms.Select_ById{
				ById: legalEntityRl.UoM2.Id,
			},
		}
	}

	if legalEntityRl.UoM3 != nil {
		msg.Currency3 = &pbUoms.Select{
			Select: &pbUoms.Select_ById{
				ById: legalEntityRl.UoM3.Id,
			},
		}
	}

	return nil
}

// Check if update to an already-existed Name
func (s *ServiceServer) validateUpdateValue(
	msg *pbLegalEntities.UpdateRequest,
	oldLegalEntity *pbLegalEntities.LegalEntity,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)

	if msg.Name != nil && msg.GetName() != oldLegalEntity.Name {
		if isUniq, errCode := s.IsLegalEntityUniq(&pbLegalEntities.Select_ByName{
			ByName: msg.GetName(),
		}); !isUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage("Name has been used")
		}
	}

	return nil
}
