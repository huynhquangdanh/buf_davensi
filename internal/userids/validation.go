package userids

import (
	"context"
	"fmt"
	"reflect"

	pbContacts "davensi.com/core/gen/contacts"
	pbUsers "davensi.com/core/gen/users"

	pbCommon "davensi.com/core/gen/common"
	pbIncomes "davensi.com/core/gen/incomes"
	pbUserIDs "davensi.com/core/gen/userids"
	"davensi.com/core/internal/common"
)

func (s *ServiceServer) validateCreate(req *pbUserIDs.CreateRequest) (errno pbCommon.ErrorCode, err error) {
	// Verify that User is specified
	if req.User == nil {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf(common.Errors[uint32(errno)], "setting "+_entityName, "user.User must be specified")
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

func (s *ServiceServer) validateUpdate(msg *pbUserIDs.UpdateRequest) (errno pbCommon.ErrorCode, err error) {
	// Verify that users.User is specified
	if msg.GetUser() == nil {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf(common.Errors[uint32(errno)], "updating "+_entityName, "user.User must be specified")
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

func (s *ServiceServer) validateModifyContact(ctx context.Context, msg any, scanUserID *string) (errno pbCommon.ErrorCode, err error) {
	var (
		userReq          *pbUsers.Select
		contactReq       *pbContacts.SetLabeledContactList
		updateContactReq *pbContacts.UpdateLabeledContactRequest
	)
	switch convertedMsg := msg.(type) {
	case *pbUserIDs.SetContactsRequest:
		userReq = convertedMsg.GetUser()
		contactReq = convertedMsg.GetContacts()
	case *pbUserIDs.AddContactsRequest:
		userReq = convertedMsg.GetUser()
		contactReq = convertedMsg.GetContacts()
	case *pbUserIDs.UpdateContactRequest:
		userReq = convertedMsg.GetUser()
		updateContactReq = convertedMsg.GetContact()
	}
	if reflect.ValueOf(userReq).IsNil() {
		return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			fmt.Errorf(common.Errors[uint32(errno)], "adding/setting contacts: user.User must be specified")
	}
	isUpdateContactRequest := func(msg any) bool {
		_, ok := msg.(*pbUserIDs.UpdateContactRequest)
		return ok
	}
	if !isUpdateContactRequest(msg) && (reflect.ValueOf(contactReq).IsNil() || len(contactReq.GetList()) == 0) {
		return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			fmt.Errorf(common.Errors[uint32(errno)], "adding/setting contacts: contacts must be specified")
	}
	if isUpdateContactRequest(msg) {
		if reflect.ValueOf(updateContactReq).IsNil() {
			return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
				fmt.Errorf(common.Errors[uint32(errno)], "updating contacts: update contact request must be specified")
		}
		if updateContactReq.GetByLabel() == "" {
			return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
				fmt.Errorf(common.Errors[uint32(errno)], "updating contacts: label must be specified")
		}
		if reflect.ValueOf(updateContactReq.GetContact()).IsNil() &&
			reflect.ValueOf(updateContactReq.MainContact).IsNil() &&
			reflect.ValueOf(updateContactReq.Status).IsNil() {
			return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
				fmt.Errorf(common.Errors[uint32(errno)], "updating contacts: either contact/main_contact/status must be specified")
		}
	}

	if userIDScanErr := s.scanForExistUser(ctx, userReq, scanUserID); userIDScanErr != nil {
		return pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR, userIDScanErr
	}
	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

func (s *ServiceServer) ValidateAddAddresses(req *pbUserIDs.AddAddressesRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.User.GetById() == "" && req.User.GetByLogin() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "user.User must be specified")
	}
	return
}

func (s *ServiceServer) ValidateSetAddresses(req *pbUserIDs.SetAddressesRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.User.GetById() == "" && req.User.GetByLogin() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "setting "+_entityName, "user.User must be specified")
	}
	return
}
func (s *ServiceServer) ValidateUpdateAddress(req *pbUserIDs.UpdateAddressRequest) (errCode pbCommon.ErrorCode, err error) {
	if req.User.GetById() == "" && req.User.GetByLogin() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "updating "+_entityName, "user.User must be specified")
		return errCode, err
	}
	if reflect.ValueOf(req.GetAddress()).IsNil() || req.GetAddress().GetByLabel() == "" {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(common.Errors[uint32(errCode)], "updating "+_entityName, "Address/ByLabel must be specified")
		return errCode, err
	}
	if reflect.ValueOf(req.GetAddress().GetAddress()).IsNil() &&
		reflect.ValueOf(req.GetAddress().Status).IsNil() &&
		reflect.ValueOf(req.GetAddress().MainAddress).IsNil() &&
		reflect.ValueOf(req.GetAddress().OwnershipStatus).IsNil() {
		errCode = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err = fmt.Errorf(
			common.Errors[uint32(errCode)],
			"updating "+_entityName, "either Address/Status/MainAddress/OwnershipStatus must be specified",
		)

		return errCode, err
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}
func (s *ServiceServer) CheckUpdateability(
	oldUserContact *pbContacts.LabeledContact,
	updateUserContactRequest *pbContacts.UpdateLabeledContactRequest,
) bool {
	isUpdateable := true
	if !reflect.ValueOf(updateUserContactRequest.MainContact).IsNil() {
		isUpdateable = oldUserContact.GetMainContact() != updateUserContactRequest.GetMainContact()
	}
	if !reflect.ValueOf(updateUserContactRequest.Status).IsNil() {
		isUpdateable = oldUserContact.GetStatus() != updateUserContactRequest.GetStatus()
	}
	if !reflect.ValueOf(updateUserContactRequest.Contact).IsNil() {
		updateContact := updateUserContactRequest.GetContact()
		if !reflect.ValueOf(updateContact.Label).IsNil() {
			isUpdateable = oldUserContact.GetLabel() != updateContact.GetLabel()
		}
		if !reflect.ValueOf(updateContact.Value).IsNil() {
			isUpdateable = oldUserContact.GetContact().GetValue() != updateContact.GetValue()
		}
		if !reflect.ValueOf(updateContact.Type).IsNil() {
			isUpdateable = oldUserContact.GetContact().GetType() != updateContact.GetType()
		}
		if !reflect.ValueOf(updateContact.Status).IsNil() {
			isUpdateable = oldUserContact.GetContact().GetStatus() != updateContact.GetStatus()
		}
	}
	return isUpdateable
}

func (s *ServiceServer) validateModifyIncome(ctx context.Context, msg any, scanUserID *string) (errno pbCommon.ErrorCode, err error) {
	var (
		userReq         *pbUsers.Select
		incomeReq       *pbIncomes.SetLabeledIncomeList
		updateIncomeReq *pbIncomes.UpdateLabeledIncomeRequest
	)
	switch convertedMsg := msg.(type) {
	case *pbUserIDs.SetIncomesRequest:
		userReq = convertedMsg.GetUser()
		incomeReq = convertedMsg.GetIncomes()
	case *pbUserIDs.AddIncomesRequest:
		userReq = convertedMsg.GetUser()
		incomeReq = convertedMsg.GetIncomes()
	}
	if reflect.ValueOf(userReq).IsNil() {
		return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			fmt.Errorf(common.Errors[uint32(errno)], "adding/setting contacts: user.User must be specified")
	}
	if _, ok := msg.(*pbUserIDs.UpdateContactRequest); !ok && (reflect.ValueOf(incomeReq).IsNil() || len(incomeReq.GetList()) == 0) {
		return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			fmt.Errorf(common.Errors[uint32(errno)], "adding/setting contacts: contacts must be specified")
	} else if ok && reflect.ValueOf(updateIncomeReq).IsNil() {
		return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			fmt.Errorf(common.Errors[uint32(errno)], "updating contacts: update contact request must be specified")
	} else if ok && updateIncomeReq.GetByLabel() == "" {
		return pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			fmt.Errorf(common.Errors[uint32(errno)], "updating contacts: label must be specified")
	}

	if userIDScanErr := s.scanForExistUser(ctx, userReq, scanUserID); userIDScanErr != nil {
		return pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR, userIDScanErr
	}
	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}
