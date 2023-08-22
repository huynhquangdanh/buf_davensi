package userids

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbKyc "davensi.com/core/gen/kyc"
	pbUserIDs "davensi.com/core/gen/userids"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/contacts"
	internalUserContact "davensi.com/core/internal/usercontacts"
	"davensi.com/core/internal/util"
)

func (s *ServiceServer) GetContacts(ctx context.Context, userID string) (*pbContacts.LabeledContactList, error) {
	rows, err := s.GetUser(ctx, userID, _contactTableName, _userContactFields)
	if err != nil {
		return nil, err
	}

	userContactsRes, err := s.userContactRepo.ScanMultiRows(rows)
	if err != nil || len(userContactsRes.GetList()) == 0 {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return nil, err
	}

	list := userContactsRes.GetList()
	newList := []*pbContacts.LabeledContact{}

	for _, c := range list {
		id := c.GetContact().GetId()
		contactRes, contactErr := s.contactSS.Get(ctx, connect.NewRequest[pbContacts.GetRequest](&pbContacts.GetRequest{Id: id}))
		if contactErr == nil {
			c.Contact = contactRes.Msg.GetContact()
			newList = append(newList, c)
		}
	}
	userContactsRes.List = newList

	return userContactsRes, nil
}

func (s *ServiceServer) RemoveContacts(
	ctx context.Context,
	req *connect.Request[pbUserIDs.RemoveContactsRequest],
) (*connect.Response[pbUserIDs.RemoveContactsResponse], error) {
	var (
		row pgx.Rows
		err error
		qb  *util.QueryBuilder
	)

	// Verify that User exists in core.users
	var userContact string
	row, err = s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
			Response: &pbUserIDs.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}

	_err := s.CheckRow(row, userContact)
	if _err != nil {
		return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
			Response: &pbUserIDs.RemoveContactsResponse_Error{
				Error: _err,
			},
		}), err
	}

	_qb := s.repo.QbGetOne(userContact, _contactTableName, _userContactFields, req.Msg.GetLabels().GetList())
	sqlstr, sqlArgs, sel := _qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
			Response: &pbUserIDs.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	defer rows.Close()

	userContactsRes, err := s.userContactRepo.ScanMultiRows(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
			Response: &pbUserIDs.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	if len(userContactsRes.GetList()) == 0 {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
			Response: &pbUserIDs.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	ids := []string{}
	list := userContactsRes.GetList()
	for _, c := range list {
		ids = append(ids, c.GetContact().GetId())
	}

	qb, err = s.repo.QbDelete(userContact, _contactTableName, req.Msg.GetLabels().GetList())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
			Response: &pbUserIDs.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}
	sqlstr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	var errScan error

	log.Info().Msg("Executing SQL '" + sqlstr + "'")
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		fmt.Println(ids)
		_qb, _err := contacts.GetSingletonServiceServer(s.db).Repo.QbDeleteMany(ids)
		if _err != nil {
			log.Error().Err(err)
			return _err
		}

		_sqlstr, _args, _ := _qb.SetReturnFields("*").GenerateSQL()
		log.Info().Msg("Executing SQL '" + _sqlstr + "'")
		if row, err = tx.Query(ctx, _sqlstr, _args...); err != nil {
			return err
		}
		row.Close()

		if rows, err = tx.Query(ctx, sqlstr, args...); err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if _, errScan = s.userContactRepo.ScanMultiRows(rows); errScan != nil {
				log.Error().Err(err).Msgf("unable to delete %s with id/login = '%s'",
					_entityName, req.Msg.GetUser().String())
				return errScan
			}
			log.Info().Msgf("%s with id/login = '%s' removed successfully",
				_entityName, req.Msg.GetUser().String())
			return nil
		} else {
			return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND)],
				_entityName, "id/login="+req.Msg.GetUser().String())
		}
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"creating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
			Response: &pbUserIDs.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserIDs.RemoveContactsResponse{
		Response: &pbUserIDs.RemoveContactsResponse_Contacts{
			Contacts: &pbUserIDs.ContactList{
				User:     req.Msg.GetUser(),
				Contacts: userContactsRes,
			},
		},
	}), nil
}

func (s *ServiceServer) AddContacts(
	ctx context.Context,
	req *connect.Request[pbUserIDs.AddContactsRequest],
) (*connect.Response[pbUserIDs.AddContactsResponse], error) {
	// Verify that User exists in core.users
	var (
		userID                         string
		insertUserContactByIDList      []*internalUserContact.ModifyUserContactParams
		insertUserContactByContactList []*internalUserContact.ModifyUserContactParams
		createByExistContactIDQB       *util.QueryBuilder
		createByNewContactQB           *util.QueryBuilder
		createNewContactQB             *util.QueryBuilder
		QBErr                          error
	)
	errorNo, errValidateAddContact := s.validateModifyContact(ctx, req.Msg, &userID)
	if errValidateAddContact != nil {
		err := fmt.Errorf(common.Errors[uint32(errorNo)],
			"adding contacts", errValidateAddContact.Error())
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.AddContactsResponse{
			Response: &pbUserIDs.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorNo,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	for _, contactReq := range req.Msg.GetContacts().GetList() {
		if contactScanErr := s.userContactSS.ProcessModifyContacts(
			ctx, contactReq, "ADD_CONTACTS", userID,
			&insertUserContactByIDList,
			&insertUserContactByContactList,
			s.contactSS,
		); contactScanErr != nil {
			return connect.NewResponse(&pbUserIDs.AddContactsResponse{
				Response: &pbUserIDs.AddContactsResponse_Error{
					Error: contactScanErr,
				},
			}), errors.New(contactScanErr.Text)
		}
	}
	createByExistContactIDQB, QBErr = s.userContactRepo.QbBulkInsert(
		userID,
		mapToSetLabeledContactsByID(insertUserContactByIDList),
	)
	if QBErr != nil {
		return connect.NewResponse(&pbUserIDs.AddContactsResponse{
			Response: &pbUserIDs.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    QBErr.Error(),
				},
			},
		}), QBErr
	}

	createByExistContactIDSQLStr, createByExistContactIDSQLArgs, _ := createByExistContactIDQB.GenerateSQL()

	newUserContactList := &pbUserIDs.ContactList{
		User: req.Msg.GetUser(),
		Contacts: &pbContacts.LabeledContactList{
			List: []*pbContacts.LabeledContact{},
		},
	}
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, labelContactByIDErr := common.TxBulkWrite[pbContacts.LabeledContact](
			ctx, tx, createByExistContactIDSQLStr, createByExistContactIDSQLArgs, s.userContactRepo.ScanLabeledContactRows)
		if labelContactByIDErr != nil {
			return labelContactByIDErr
		}
		createContactRequests := mapToCreateContactRequests(insertUserContactByContactList)
		// make a slice of uuid string equal to the len of createContactRequests
		uuidSlice := make([]string, len(createContactRequests))
		for i := range createContactRequests {
			uuidSlice[i] = uuid.New().String()
		}
		createNewContactQB, QBErr = s.contactRepo.QbBulkInsertMany(createContactRequests, uuidSlice)
		if QBErr != nil {
			return QBErr
		}
		createContactsSQLStr, createContactsSQLArgs, _ := createNewContactQB.GenerateSQL()
		_, createContactErr := common.TxBulkWrite[pbContacts.Contact](
			ctx, tx, createContactsSQLStr, createContactsSQLArgs, s.contactRepo.ScanMultiRows)
		if createContactErr != nil {
			return createContactErr
		}
		createByNewContactQB, QBErr = s.userContactRepo.QbBulkInsert(
			userID,
			mapToSetLabeledContactsByID(insertUserContactByContactList),
		)
		if QBErr != nil {
			return QBErr
		}
		createByNewContactSQLStr, createByNewContactSQLArgs, _ := createByNewContactQB.GenerateSQL()
		_, labelContactByNewContactErr := common.TxBulkWrite[pbContacts.LabeledContact](
			ctx, tx, createByNewContactSQLStr, createByNewContactSQLArgs, s.userContactRepo.ScanLabeledContactRows)
		if labelContactByNewContactErr != nil {
			return labelContactByNewContactErr
		}
		// merge two slice insertUserContactByContactList and insertUserContactByIDList
		newUserContactList.Contacts.List = mapToLabelContactList(append(insertUserContactByContactList, insertUserContactByIDList...))

		return nil
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"remove", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.AddContactsResponse{
			Response: &pbUserIDs.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}
	return connect.NewResponse(&pbUserIDs.AddContactsResponse{
		Response: &pbUserIDs.AddContactsResponse_Contacts{
			Contacts: newUserContactList,
		},
	}), nil
}

func mapToSetLabeledContactsByID(paramsList []*internalUserContact.ModifyUserContactParams) []*pbContacts.SetLabeledContact {
	var setLabeledContacts []*pbContacts.SetLabeledContact

	// Loop through each *ModifyUserContactParams and create corresponding *SetLabeledContact
	for _, param := range paramsList {
		setLabeledContact := &pbContacts.SetLabeledContact{
			Label: param.Label,
			Select: &pbContacts.SetLabeledContact_Id{
				Id: param.Contact.ID,
			},
			Status:      param.Status,
			MainContact: &param.MainContact,
		}

		setLabeledContacts = append(setLabeledContacts, setLabeledContact)
	}

	return setLabeledContacts
}

func mapToCreateContactRequests(paramsList []*internalUserContact.ModifyUserContactParams) []*pbContacts.CreateRequest {
	var createContactRequests []*pbContacts.CreateRequest

	// Loop through each *ModifyUserContactParams and create corresponding *pbContacts.CreateRequest
	for _, param := range paramsList {
		createContactReq := &pbContacts.CreateRequest{
			Type:   param.Contact.Type,
			Value:  param.Contact.Value,
			Status: param.Contact.Status,
		}
		createContactRequests = append(createContactRequests, createContactReq)
	}

	return createContactRequests
}

func mapToLabelContactList(paramsList []*internalUserContact.ModifyUserContactParams) []*pbContacts.LabeledContact {
	var labelContactList []*pbContacts.LabeledContact
	for _, param := range paramsList {
		labelContact := &pbContacts.LabeledContact{
			Label: param.Label,
			Contact: &pbContacts.Contact{
				Type:   param.Contact.Type,
				Value:  param.Contact.Value,
				Status: *param.Contact.Status.Enum(),
				Id:     param.Contact.ID,
			},
			Status:      *param.Status.Enum(),
			MainContact: &param.MainContact,
		}
		labelContactList = append(labelContactList, labelContact)
	}

	return labelContactList
}

func (s *ServiceServer) InsertManyUserContactByIDs(
	ctx context.Context,
	insertUserContactByIDList []*internalUserContact.ModifyUserContactParams,
	scanUserID string,
	tx pgx.Tx,
) (err error) {
	insertByExistContactIDQB, err := s.userContactRepo.QbBulkInsert(
		scanUserID,
		mapToSetLabeledContactsByID(insertUserContactByIDList),
	)
	if err != nil {
		return
	}
	insertByExistContactIDSQLStr, insertByExistContactIDSQLArgs, _ := insertByExistContactIDQB.GenerateSQL()
	_, err = common.TxBulkWrite[pbContacts.LabeledContact](
		ctx, tx, insertByExistContactIDSQLStr, insertByExistContactIDSQLArgs, s.userContactRepo.ScanLabeledContactRows)
	if err != nil {
		return
	}
	return
}

func (s *ServiceServer) UpdateManyUserContactByIDs(
	ctx context.Context, updateUserContactByIDList []*internalUserContact.ModifyUserContactParams,
	scanUserID string, tx pgx.Tx,
) (err error) {
	updateUserContactQBs, err := s.repo.QbUserAdditionInfoUpdateMany(
		_contactTableName,
		scanUserID,
		SupportedValueSlices{
			UserContactValueList: mapToSetLabeledContactsByID(updateUserContactByIDList),
		},
	)
	if err != nil {
		return
	}
	// merge updateUserContactQBs into one query
	updateBatch := &pgx.Batch{}
	for _, qb := range updateUserContactQBs {
		sqlStr, args, _ := qb.GenerateSQL()
		updateBatch.Queue(sqlStr, args...)
	}
	_, err = common.TxSendBatch[pbContacts.LabeledContact](
		ctx, tx, updateBatch, len(updateUserContactQBs), s.userContactRepo.ScanRow, nil, true,
	)
	if err != nil {
		return
	}
	return
}

func (s *ServiceServer) UpsertUserContactByContactList(
	ctx context.Context,
	upsertUserContactByContactList []*internalUserContact.ModifyUserContactParams,
	scanUserID string, tx pgx.Tx,
) error {
	createContactRequests := mapToCreateContactRequests(upsertUserContactByContactList)
	// make a slice of uuid string equal to the len of createContactRequests
	uuidSlice := make([]string, len(createContactRequests))
	for i := range createContactRequests {
		uuidSlice[i] = uuid.New().String()
	}
	createNewContactQB, err := s.contactRepo.QbBulkInsertMany(createContactRequests, uuidSlice)
	if err != nil {
		return err
	}
	createContactsSQLStr, createContactsSQLArgs, _ := createNewContactQB.GenerateSQL()
	_, err = common.TxBulkWrite[pbContacts.Contact](
		ctx, tx, createContactsSQLStr, createContactsSQLArgs, s.contactSS.Repo.ScanMultiRows)
	if err != nil {
		return err
	}
	createNewContactQB, err = s.userContactRepo.QbBulkInsert(
		scanUserID,
		mapToSetLabeledContactsByID(upsertUserContactByContactList),
	)
	if err != nil {
		return err
	}
	upsertByNewContactSQLStr, upsertByNewContactSQLArgs, _ := createNewContactQB.GenerateSQL()
	if _, err := common.TxBulkWrite[pbContacts.LabeledContact](
		ctx, tx, upsertByNewContactSQLStr, upsertByNewContactSQLArgs, s.userContactRepo.ScanLabeledContactRows); err != nil {
		return err
	}
	return nil
}

func (s *ServiceServer) SetContacts(
	ctx context.Context,
	req *connect.Request[pbUserIDs.SetContactsRequest],
) (*connect.Response[pbUserIDs.SetContactsResponse], error) {
	// Verify that User exists in core.users
	var (
		scanUserID                     string
		upsertUserContactByIDList      []*internalUserContact.ModifyUserContactParams
		upsertUserContactByContactList []*internalUserContact.ModifyUserContactParams
	)
	if errorNo, validateSetContactErr := s.validateModifyContact(ctx, req.Msg, &scanUserID); validateSetContactErr != nil {
		err := fmt.Errorf(common.Errors[uint32(errorNo)], "setting contacts", validateSetContactErr.Error())
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.SetContactsResponse{
			Response: &pbUserIDs.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorNo,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	// filter request to get insert and update contacts

	for _, contactReq := range req.Msg.GetContacts().GetList() {
		if contactScanErr := s.userContactSS.ProcessModifyContacts(
			ctx, contactReq, "UPSERT_CONTACTS", scanUserID,
			&upsertUserContactByIDList,
			&upsertUserContactByContactList,
			s.contactSS,
		); contactScanErr != nil {
			return connect.NewResponse(&pbUserIDs.SetContactsResponse{
				Response: &pbUserIDs.SetContactsResponse_Error{
					Error: contactScanErr,
				},
			}), errors.New(contactScanErr.Text)
		}
	}
	insertUserContactByIDList := util.Filter(upsertUserContactByIDList, func(elem *internalUserContact.ModifyUserContactParams) bool {
		return elem.ModifyType == "ADD_CONTACTS"
	})
	updateUserContactByIDList := util.Filter(upsertUserContactByIDList, func(elem *internalUserContact.ModifyUserContactParams) bool {
		return elem.ModifyType == "UPDATE_CONTACTS"
	})

	upsertedUserContactList := &pbUserIDs.ContactList{
		User: req.Msg.GetUser(),
		Contacts: &pbContacts.LabeledContactList{
			List: []*pbContacts.LabeledContact{},
		},
	}
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if len(insertUserContactByIDList) > 0 {
			if err := s.InsertManyUserContactByIDs(ctx, insertUserContactByIDList, scanUserID, tx); err != nil {
				return err
			}
			upsertedUserContactList.Contacts.List = append(
				upsertedUserContactList.Contacts.List,
				mapToLabelContactList(insertUserContactByIDList)...,
			)
		}
		if len(updateUserContactByIDList) > 0 {
			if err := s.UpdateManyUserContactByIDs(ctx, updateUserContactByIDList, scanUserID, tx); err != nil {
				return err
			}

			upsertedUserContactList.Contacts.List = append(
				upsertedUserContactList.Contacts.List,
				mapToLabelContactList(updateUserContactByIDList)...,
			)
		}
		if len(upsertUserContactByContactList) > 0 {
			if err := s.UpsertUserContactByContactList(ctx, upsertUserContactByContactList, scanUserID, tx); err != nil {
				return err
			}
			upsertedUserContactList.Contacts.List = append(
				upsertedUserContactList.Contacts.List,
				mapToLabelContactList(upsertUserContactByContactList)...,
			)
		}

		return nil
	}); errTx != nil {
		log.Error().Err(errTx).Msg(errTx.Error())
		return connect.NewResponse(&pbUserIDs.SetContactsResponse{
			Response: &pbUserIDs.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    errTx.Error(),
				},
			},
		}), errTx
	}
	return connect.NewResponse(&pbUserIDs.SetContactsResponse{
		Response: &pbUserIDs.SetContactsResponse_Contacts{
			Contacts: upsertedUserContactList,
		},
	}), nil
}

func (s *ServiceServer) GetUserContactByIDLabel(
	ctx context.Context,
	userID, label string,
) (*pbContacts.LabeledContact, error) {
	qb := s.repo.QbGetOne(userID, _contactTableName, _userContactFields, []string{label})
	sqlStr, sqlArgs, _ := qb.GenerateSQL()
	queryRows, queryErr := s.db.Query(ctx, sqlStr, sqlArgs...)
	if queryErr != nil {
		return nil, queryErr
	}

	defer queryRows.Close()

	rowErr := queryRows.Err()

	fmt.Println(rowErr, sqlStr, sqlArgs)

	if rowErr != nil {
		return nil, rowErr
	}

	if !queryRows.Next() {
		return nil, errors.New("user contact with user_id: " + userID + " and label: " + label + " not found")
	}
	userContact, userContactScanErr := s.userContactRepo.ScanLabelContactSingleRow(queryRows)
	if userContactScanErr != nil {
		return nil, userContactScanErr
	}
	if userContact != nil &&
		(userContact.GetStatus() == pbKyc.Status_STATUS_UNSPECIFIED ||
			userContact.GetStatus() == pbKyc.Status_STATUS_CANCELED) {
		return nil, errors.New("user contact with user_id: " + userID + " and label: " + label + "is deleted or not active")
	}
	// get old contact detail
	contactDetail, getContactDetailErr := s.contactSS.Get(ctx, &connect.Request[pbContacts.GetRequest]{
		Msg: &pbContacts.GetRequest{
			Id: userContact.GetContact().GetId(),
		},
	})
	if getContactDetailErr != nil {
		return nil, getContactDetailErr
	}
	userContact.Contact = contactDetail.Msg.GetContact()

	return userContact, nil
}

func (s *ServiceServer) UpdateContact(
	ctx context.Context,
	req *connect.Request[pbUserIDs.UpdateContactRequest],
) (*connect.Response[pbUserIDs.UpdateContactResponse], error) {
	var (
		scanUserID           string
		updateUserContractQB *util.QueryBuilder
		updateContactQB      *util.QueryBuilder
		QBErr                error
	)
	errorNo, validateUpdateContactErr := s.validateModifyContact(ctx, req.Msg, &scanUserID)
	if validateUpdateContactErr != nil {
		err := fmt.Errorf(common.Errors[uint32(errorNo)],
			"setting contacts", validateUpdateContactErr.Error())
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.UpdateContactResponse{
			Response: &pbUserIDs.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorNo,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	// get old user contact to update
	oldUserContact, oldUserContactErr := s.GetUserContactByIDLabel(ctx, scanUserID, req.Msg.GetContact().GetByLabel())
	if oldUserContactErr != nil {
		err := fmt.Errorf(common.Errors[uint32(errorNo)],
			"updating contacts", oldUserContactErr.Error())
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.UpdateContactResponse{
			Response: &pbUserIDs.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorNo,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	if isUpdateable := s.CheckUpdateability(oldUserContact, req.Msg.GetContact()); !isUpdateable {
		return connect.NewResponse(&pbUserIDs.UpdateContactResponse{
			Response: &pbUserIDs.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    "contact is not updateable with provided data",
				},
			},
		}), errors.New("contact is not updateable with provided data")
	}

	updateUserContractQB, QBErr = s.repo.QbUpdate(scanUserID, pbKyc.Status_STATUS_UNSPECIFIED, _contactTableName, req.Msg.GetContact())
	if QBErr != nil {
		return connect.NewResponse(&pbUserIDs.UpdateContactResponse{
			Response: &pbUserIDs.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    QBErr.Error(),
				},
			},
		}), QBErr
	}
	if !reflect.ValueOf(req.Msg.GetContact().GetContact()).IsNil() &&
		(!reflect.ValueOf(req.Msg.GetContact().GetContact().Value).IsNil() ||
			!reflect.ValueOf(req.Msg.GetContact().GetContact().Status).IsNil() ||
			!reflect.ValueOf(req.Msg.GetContact().GetContact().Type).IsNil()) {
		updateContactQB, QBErr = s.contactRepo.QbUpdate(&pbContacts.UpdateRequest{
			Id:     oldUserContact.GetContact().GetId(),
			Value:  req.Msg.GetContact().GetContact().Value,
			Status: req.Msg.GetContact().GetContact().Status,
			Type:   req.Msg.GetContact().GetContact().Type,
		})
		if QBErr != nil {
			return connect.NewResponse(&pbUserIDs.UpdateContactResponse{
				Response: &pbUserIDs.UpdateContactResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    QBErr.Error(),
					},
				},
			}), QBErr
		}
	}
	updateUserContractSQLStr, updateUserContractSQLArgs, _ := updateUserContractQB.GenerateSQL()
	updateUserContact := &pbUserIDs.Contact{
		User:    req.Msg.GetUser(),
		Contact: &pbContacts.LabeledContact{},
	}
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		updatedlabelContact, updatelabelContactErr := common.TxWrite[pbContacts.LabeledContact](
			ctx, tx, updateUserContractSQLStr, updateUserContractSQLArgs, s.userContactRepo.ScanLabelContactSingleRow)
		if updatelabelContactErr != nil {
			return updatelabelContactErr
		}
		updateUserContact.Contact = updatedlabelContact
		updateUserContact.Contact.Contact = oldUserContact.GetContact()
		if !reflect.ValueOf(updateContactQB).IsNil() {
			updateContractSQLStr, updateContractSQLArgs, _ := updateContactQB.GenerateSQL()
			updatedContact, updateContractErr := common.TxWrite[pbContacts.Contact](
				ctx, tx, updateContractSQLStr, updateContractSQLArgs, s.contactRepo.ScanRow)
			if updateContractErr != nil {
				return updateContractErr
			}
			updateUserContact.Contact.Contact = updatedContact
		}
		return nil
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"creating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.UpdateContactResponse{
			Response: &pbUserIDs.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}
	return connect.NewResponse(&pbUserIDs.UpdateContactResponse{
		Response: &pbUserIDs.UpdateContactResponse_Contact{
			Contact: updateUserContact,
		},
	}), nil
}
