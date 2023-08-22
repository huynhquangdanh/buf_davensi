package cexaccounts

import (
	"errors"

	pbCexAccount "davensi.com/core/gen/cexaccounts"
)

func (s *ServiceServer) validateCreateRequest(_ *pbCexAccount.CreateRequest) error {
	return nil
}

func (s *ServiceServer) ValidateUpdateRequest(req *pbCexAccount.UpdateRequest) (err error) {
	err = nil
	if req.GetRecipient().GetSelect().GetById() != "" {
		return
	} else if req.GetRecipient().GetSelect().GetByLegalEntityUserLabel() != nil {
		return
	}
	err = errors.New("id or legal entity user label must be specified")
	return
}
