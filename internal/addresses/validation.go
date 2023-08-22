package addresses

import (
	pbAddresses "davensi.com/core/gen/addresses"
	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/countries"
)

// for Update gRPC
// func validateQueryUpdate(msg *pbAddresses.UpdateRequest) error {
// 	// Verify that ID is specified
// 	if msg.GetId() == "" {
// 		return errors.New("id must be specified")
// 	}

// 	return nil
// }

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbAddresses.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	if errSelectCountry := countries.ValidateSelect(msg.GetCountry(), "creating"); errSelectCountry != nil {
		return errSelectCountry
	}

	addressRl := s.GetRelationship(msg.GetCountry())

	if addressRl.Country == nil {
		return errCreation.UpdateMessage("country does not exist")
	}

	msg.Country = &pbCountries.Select{
		Select: &pbCountries.Select_ById{
			ById: addressRl.Country.Id,
		},
	}

	return nil
}

// for Update gRPC
// Check for WHERE clause, whether the relationships exist
func (s *ServiceServer) validateUpdateQuery(msg *pbAddresses.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	addressRl := s.GetRelationship(msg.GetCountry())

	if msg.Country != nil && addressRl.Country == nil {
		return errUpdate.UpdateMessage("country does not exist")
	}

	if addressRl.Country != nil {
		msg.Country = &pbCountries.Select{
			Select: &pbCountries.Select_ById{
				ById: addressRl.Country.Id,
			},
		}
	}

	return nil
}
