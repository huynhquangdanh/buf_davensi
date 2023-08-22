package credentials

import (
	"errors"

	pbCommon "davensi.com/core/gen/common"
	pbCredentials "davensi.com/core/gen/credentials"
)

func (s *ServiceServer) validateQueryUpdate(msg *pbCredentials.UpdateRequest) error {
	// Verify that ID is specified
	if msg.GetId() == "" {
		return errors.New("id must be specified")
	}
	return nil
}

// validator query insert when insert into database
func (s *ServiceServer) validateQueryInsert(_ *pbCredentials.CreateRequest) error {
	// TODO: sangly validate birthday, country_of_birth, country_of_nationality later
	return nil
}

func (s *ServiceServer) validateUpdateValue(_ *pbCredentials.Credentials, _ *pbCredentials.UpdateRequest) (pbCommon.ErrorCode, error) {
	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

func (s *ServiceServer) validateMsgGetOne(msg *pbCredentials.GetRequest) (err error) {
	err = nil
	if msg.GetId() == "" {
		err = errors.New("id must be specified")
	}
	return
}
