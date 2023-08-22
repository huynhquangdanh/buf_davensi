package userids

import (
	"context"
	"errors"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbIncomes "davensi.com/core/gen/incomes"
	pbUserIDs "davensi.com/core/gen/userids"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/incomes"
	internalUserIncome "davensi.com/core/internal/userincomes"
	"davensi.com/core/internal/util"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func (s *ServiceServer) GetIncomes(ctx context.Context, userID string) (*pbIncomes.LabeledIncomeList, error) {
	qb := s.repo.QbGetOne(userID, _incomeTableName, _userIncomeFields, []string{})
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return nil, err
	}

	userIncomesRes, err := s.userIncomeRepo.ScanMultiRows(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return nil, err
	}
	if len(userIncomesRes.GetList()) == 0 {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return nil, _err
	}

	list := userIncomesRes.GetList()
	newList := []*pbIncomes.LabeledIncome{}

	for _, c := range list {
		id := c.GetIncome().GetId()
		incomeRes, incomeErr := s.incomeSS.Get(ctx, connect.NewRequest[pbIncomes.GetRequest](&pbIncomes.GetRequest{Id: id}))
		if incomeErr == nil {
			c.Income = incomeRes.Msg.GetIncome()
			newList = append(newList, c)
		}
	}
	userIncomesRes.List = newList

	return userIncomesRes, nil
}

func (s *ServiceServer) RemoveIncomes(
	ctx context.Context,
	req *connect.Request[pbUserIDs.RemoveIncomesRequest],
) (*connect.Response[pbUserIDs.RemoveIncomesResponse], error) {
	var (
		row pgx.Rows
		err error
		qb  *util.QueryBuilder
	)

	// Verify that User exists in core.users
	var userIncome string
	row, err = s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
			Response: &pbUserIDs.RemoveIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _incomePackage,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}

	_err := s.CheckRow(row, userIncome)
	if _err != nil {
		return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
			Response: &pbUserIDs.RemoveIncomesResponse_Error{
				Error: _err,
			},
		}), err
	}

	_qb := s.repo.QbGetOne(userIncome, _incomeTableName, _userIncomeFields, req.Msg.GetLabels().GetList())
	sqlstr, sqlArgs, sel := _qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
			Response: &pbUserIDs.RemoveIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _incomePackage,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	defer rows.Close()

	userIncomesRes, err := s.userIncomeRepo.ScanMultiRows(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
			Response: &pbUserIDs.RemoveIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _incomePackage,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	if len(userIncomesRes.GetList()) == 0 {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
			Response: &pbUserIDs.RemoveIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _incomePackage,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	ids := []string{}
	list := userIncomesRes.GetList()
	for _, c := range list {
		ids = append(ids, c.GetIncome().GetId())
	}

	qb, err = s.repo.QbDelete(userIncome, _incomeTableName, req.Msg.GetLabels().GetList())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
			Response: &pbUserIDs.RemoveIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _incomePackage,
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
		_qb, _err := incomes.GetSingletonServiceServer(s.db).Repo.QbDeleteMany(ids)
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
			if _, errScan = s.userIncomeRepo.ScanMultiRows(rows); errScan != nil {
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
			"remove", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
			Response: &pbUserIDs.RemoveIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _incomePackage,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserIDs.RemoveIncomesResponse{
		Response: &pbUserIDs.RemoveIncomesResponse_Incomes{
			Incomes: &pbUserIDs.IncomeList{
				User:    req.Msg.GetUser(),
				Incomes: userIncomesRes,
			},
		},
	}), nil
}

func (s *ServiceServer) AddIncomes(
	ctx context.Context,
	req *connect.Request[pbUserIDs.AddIncomesRequest],
) (*connect.Response[pbUserIDs.AddIncomesResponse], error) {
	// Verify that User exists in core.users
	var (
		userID                        string
		insertUserIncomesByIDList     []*internalUserIncome.ModifyUserIncomeParams
		insertUserIncomesByIncomeList []*internalUserIncome.ModifyUserIncomeParams
		createNewIncomeMergedArr      []*internalUserIncome.ModifyUserIncomeParams
	)
	errorNo, errValidateAddContact := s.validateModifyIncome(ctx, req.Msg, &userID)
	if errValidateAddContact != nil {
		err := fmt.Errorf(common.Errors[uint32(errorNo)],
			"adding contacts", errValidateAddContact.Error())
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.AddIncomesResponse{
			Response: &pbUserIDs.AddIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorNo,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	for _, incomeReq := range req.Msg.GetIncomes().GetList() {
		if incomeScanErr := s.userIncomeSS.ProcessModifyIncomes(
			ctx, incomeReq, _incomeAddSymbol, userID,
			&insertUserIncomesByIDList,
			&insertUserIncomesByIncomeList,
			s.incomeSS,
		); incomeScanErr != nil {
			return connect.NewResponse(&pbUserIDs.AddIncomesResponse{
				Response: &pbUserIDs.AddIncomesResponse_Error{
					Error: incomeScanErr,
				},
			}), errors.New(incomeScanErr.Text)
		}
	}

	newUserIncomeList := &pbUserIDs.IncomeList{
		User: req.Msg.GetUser(),
		Incomes: &pbIncomes.LabeledIncomeList{
			List: []*pbIncomes.LabeledIncome{},
		},
	}

	tx, beginTxErr := s.db.Begin(ctx)
	if beginTxErr != nil {
		return connect.NewResponse(&pbUserIDs.AddIncomesResponse{
			Response: &pbUserIDs.AddIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    beginTxErr.Error(),
				},
			},
		}), beginTxErr
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			return
		}
	}(tx, ctx)
	if len(insertUserIncomesByIDList) > 0 {
		if err := s.InsertManyUserIncomesByIDs(ctx, insertUserIncomesByIDList, userID, tx); err != nil {
			return connect.NewResponse(&pbUserIDs.AddIncomesResponse{
				Response: &pbUserIDs.AddIncomesResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), err
		}
		newUserIncomeList.Incomes.List = append(
			newUserIncomeList.Incomes.List,
			util.MapTToR(
				insertUserIncomesByIDList,
				func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.LabeledIncome {
					return &pbIncomes.LabeledIncome{
						Label:  elem.Label,
						Income: elem.Income,
						Status: elem.Status,
					}
				})...,
		)
	}
	if len(insertUserIncomesByIncomeList) > 0 {
		if err := s.InsertUserContactByIncomeList(ctx, insertUserIncomesByIncomeList, &createNewIncomeMergedArr, userID, tx); err != nil {
			return connect.NewResponse(&pbUserIDs.AddIncomesResponse{
				Response: &pbUserIDs.AddIncomesResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), err
		}

		newUserIncomeList.Incomes.List = append(
			newUserIncomeList.Incomes.List,
			util.MapTToR(
				createNewIncomeMergedArr,
				func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.LabeledIncome {
					return &pbIncomes.LabeledIncome{
						Label:  elem.Label,
						Income: elem.Income,
						Status: elem.Status,
					}
				})...,
		)
	}

	commitTxErr := tx.Commit(ctx)
	if commitTxErr != nil {
		return connect.NewResponse(&pbUserIDs.AddIncomesResponse{
			Response: &pbUserIDs.AddIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    commitTxErr.Error(),
				},
			},
		}), commitTxErr
	}
	return connect.NewResponse(&pbUserIDs.AddIncomesResponse{
		Response: &pbUserIDs.AddIncomesResponse_Incomes{
			Incomes: newUserIncomeList,
		},
	}), nil
}
func (s *ServiceServer) InsertManyUserIncomesByIDs(
	ctx context.Context,
	insertUserIncomesByIDList []*internalUserIncome.ModifyUserIncomeParams,
	scanUserID string,
	tx pgx.Tx,
) (err error) {
	createByExistIncomeIDQB, err := s.userIncomeRepo.QbBulkInsert(
		scanUserID,
		util.MapTToR(insertUserIncomesByIDList, func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.SetLabeledIncome {
			return &pbIncomes.SetLabeledIncome{
				Label: elem.Label,
				Response: &pbIncomes.SetLabeledIncome_Id{
					Id: elem.Income.GetId(),
				},
				Status: elem.Status,
			}
		}),
	)
	if err != nil {
		return
	}

	createByExistIncomeIDSQLStr, createByExistIncomeIDSQLArgs, _ := createByExistIncomeIDQB.GenerateSQL()
	_, err = common.TxBulkWrite[pbIncomes.LabeledIncome](
		ctx, tx, createByExistIncomeIDSQLStr, createByExistIncomeIDSQLArgs, s.userIncomeRepo.ScanMultiLabelRows)

	if err != nil {
		return
	}
	return
}
func (s *ServiceServer) UpdateManyUserIncomesByIDs(
	ctx context.Context,
	updateUserIncomesByIDList []*internalUserIncome.ModifyUserIncomeParams,
	scanUserID string,
	tx pgx.Tx,
) (err error) {
	updateUserIncomeQBs, err := s.userIncomeRepo.QbUserAdditionInfoUpdateMany(
		scanUserID,
		util.MapTToR(updateUserIncomesByIDList, func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.SetLabeledIncome {
			return &pbIncomes.SetLabeledIncome{
				Label: elem.Label,
				Response: &pbIncomes.SetLabeledIncome_Id{
					Id: elem.Income.GetId(),
				},
				Status: elem.Status,
			}
		}),
	)
	if err != nil {
		return err
	}
	var updateUserIncomeBatch []*common.QueryStatements
	for _, qb := range updateUserIncomeQBs {
		sqlStr, args, _ := qb.GenerateSQL()
		updateUserIncomeBatch = append(
			updateUserIncomeBatch,
			&common.QueryStatements{
				SQLStr:  sqlStr,
				SQLArgs: args,
			})
	}
	_, err = common.TxMultipleStatement[pbIncomes.LabeledIncome](
		ctx, tx, updateUserIncomeBatch, s.userIncomeRepo.ScanMultiLabelRows,
	)
	if err != nil {
		return err
	}
	return nil
}
func (s *ServiceServer) InsertUserContactByIncomeList(
	ctx context.Context,
	insertUserIncomeByIncomeList []*internalUserIncome.ModifyUserIncomeParams,
	createNewIncomeMergedArr *[]*internalUserIncome.ModifyUserIncomeParams,
	scanUserID string, tx pgx.Tx,
) (err error) {
	createNewIncomeQBList, err := s.incomeRepo.QbBulkInsert(
		util.MapTToR(insertUserIncomeByIncomeList, func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.CreateRequest {
			return elem.CreateNewIncomeRequest
		}),
		util.MapTToR(insertUserIncomeByIncomeList, func(elem *internalUserIncome.ModifyUserIncomeParams, index int) string {
			return elem.CreateNewIncomeID
		}),
	)
	if err != nil {
		return err
	}
	// merge createNewIncomeQBList into one query
	createNewIncomeBatch := &pgx.Batch{}
	for _, qb := range createNewIncomeQBList {
		sqlStr, args, _ := qb.GenerateSQL()
		createNewIncomeBatch.Queue(sqlStr, args...)
	}
	createNewIncomeResult, resultErr := common.TxSendBatch[pbIncomes.Income](
		ctx, tx, createNewIncomeBatch, len(createNewIncomeQBList), nil, s.incomeRepo.ScanRow, false,
	)
	if resultErr != nil {
		return resultErr
	}
	// merge new createIncome ID in to insertUserContactByIncomeList
	createNewIncomeMergedMap := make(map[string]*internalUserIncome.ModifyUserIncomeParams)
	util.ForEach(insertUserIncomeByIncomeList, func(elem *internalUserIncome.ModifyUserIncomeParams, index int) {
		createNewIncomeMergedMap[elem.CreateNewIncomeID] = elem
	})
	util.ForEach(createNewIncomeResult, func(elem *pbIncomes.Income, index int) {
		if mergedEl, ok := createNewIncomeMergedMap[elem.GetId()]; ok {
			mergedEl.Income = elem
			createNewIncomeMergedMap[elem.GetId()] = mergedEl
		}
	})
	*createNewIncomeMergedArr = make([]*internalUserIncome.ModifyUserIncomeParams, 0, len(createNewIncomeMergedMap))
	for _, el := range createNewIncomeMergedMap {
		*createNewIncomeMergedArr = append(*createNewIncomeMergedArr, el)
	}
	createByNewIncomeQB, err := s.userIncomeRepo.QbBulkInsert(
		scanUserID,
		util.MapTToR(*createNewIncomeMergedArr, func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.SetLabeledIncome {
			return &pbIncomes.SetLabeledIncome{
				Label: elem.Label,
				Response: &pbIncomes.SetLabeledIncome_Id{
					Id: elem.Income.GetId(),
				},
				Status: elem.Status,
			}
		}),
	)
	if err != nil {
		return err
	}
	createByNewIncomeSQLStr, createByNewIncomeSQLArgs, _ := createByNewIncomeQB.GenerateSQL()
	_, err = common.TxBulkWrite[pbIncomes.LabeledIncome](
		ctx, tx, createByNewIncomeSQLStr, createByNewIncomeSQLArgs, s.userIncomeRepo.ScanMultiLabelRows)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServiceServer) SetIncomes(
	ctx context.Context,
	req *connect.Request[pbUserIDs.SetIncomesRequest],
) (*connect.Response[pbUserIDs.SetIncomesResponse], error) {
	// Verify that User exists in core.users
	var (
		scanUserID                   string
		upsertUserIncomeByIDList     []*internalUserIncome.ModifyUserIncomeParams
		upsertUserIncomeByIncomeList []*internalUserIncome.ModifyUserIncomeParams
		upsertNewIncomeMergedArr     []*internalUserIncome.ModifyUserIncomeParams
	)
	errorNo, validateSetContactErr := s.validateModifyIncome(ctx, req.Msg, &scanUserID)
	if validateSetContactErr != nil {
		err := fmt.Errorf(common.Errors[uint32(errorNo)],
			"setting contacts", validateSetContactErr.Error())
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.SetIncomesResponse{
			Response: &pbUserIDs.SetIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorNo,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	// filter request to get insert and update contacts

	for _, incomeReq := range req.Msg.GetIncomes().GetList() {
		if incomeScanErr := s.userIncomeSS.ProcessModifyIncomes(
			ctx, incomeReq, _incomeUpsertSymbol, scanUserID,
			&upsertUserIncomeByIDList,
			&upsertUserIncomeByIncomeList,
			s.incomeSS,
		); incomeScanErr != nil {
			return connect.NewResponse(&pbUserIDs.SetIncomesResponse{
				Response: &pbUserIDs.SetIncomesResponse_Error{
					Error: incomeScanErr,
				},
			}), errors.New(incomeScanErr.Text)
		}
	}
	insertUserIncomesByIDList := util.Filter(upsertUserIncomeByIDList, func(elem *internalUserIncome.ModifyUserIncomeParams) bool {
		return elem.ModifyType == _incomeAddSymbol
	})
	updateUserIncomesByIDList := util.Filter(upsertUserIncomeByIDList, func(elem *internalUserIncome.ModifyUserIncomeParams) bool {
		return elem.ModifyType == _incomeUpdateSymbol
	})

	upsertedUserIncomeList := &pbUserIDs.IncomeList{
		User: req.Msg.GetUser(),
		Incomes: &pbIncomes.LabeledIncomeList{
			List: []*pbIncomes.LabeledIncome{},
		},
	}

	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if len(insertUserIncomesByIDList) > 0 {
			if err := s.InsertManyUserIncomesByIDs(ctx, insertUserIncomesByIDList, scanUserID, tx); err != nil {
				return err
			}
			upsertedUserIncomeList.Incomes.List = append(
				upsertedUserIncomeList.Incomes.List,
				util.MapTToR(
					insertUserIncomesByIDList,
					func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.LabeledIncome {
						return &pbIncomes.LabeledIncome{
							Label:  elem.Label,
							Income: elem.Income,
							Status: elem.Status,
						}
					})...,
			)
		}

		if len(updateUserIncomesByIDList) > 0 {
			if err := s.UpdateManyUserIncomesByIDs(ctx, updateUserIncomesByIDList, scanUserID, tx); err != nil {
				return err
			}
			upsertedUserIncomeList.Incomes.List = append(
				upsertedUserIncomeList.Incomes.List,
				util.MapTToR(updateUserIncomesByIDList, func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.LabeledIncome {
					return &pbIncomes.LabeledIncome{
						Label:  elem.Label,
						Income: elem.Income,
						Status: elem.Status,
					}
				})...,
			)
		}
		if len(upsertUserIncomeByIncomeList) > 0 {
			if err := s.InsertUserContactByIncomeList(ctx, upsertUserIncomeByIncomeList, &upsertNewIncomeMergedArr, scanUserID, tx); err != nil {
				return err
			}
			upsertedUserIncomeList.Incomes.List = append(
				upsertedUserIncomeList.Incomes.List,
				util.MapTToR(
					upsertNewIncomeMergedArr,
					func(elem *internalUserIncome.ModifyUserIncomeParams, index int) *pbIncomes.LabeledIncome {
						return &pbIncomes.LabeledIncome{
							Label:  elem.Label,
							Income: elem.Income,
							Status: elem.Status,
						}
					})...,
			)
		}
		return nil
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"creating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.SetIncomesResponse{
			Response: &pbUserIDs.SetIncomesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}
	return connect.NewResponse(&pbUserIDs.SetIncomesResponse{
		Response: &pbUserIDs.SetIncomesResponse_Incomes{
			Incomes: upsertedUserIncomeList,
		},
	}), nil
}

func (s *ServiceServer) UpdateIncome(
	ctx context.Context,
	req *connect.Request[pbUserIDs.UpdateIncomeRequest],
) (*connect.Response[pbUserIDs.UpdateIncomeResponse], error) {
	var userIncome string
	row, err := s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
			Response: &pbUserIDs.UpdateIncomeResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _incomePackage,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}
	_err := s.CheckRow(row, userIncome)
	if _err != nil {
		return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
			Response: &pbUserIDs.UpdateIncomeResponse_Error{
				Error: _err,
			},
		}), err
	}
	qb := s.userIncomeRepo.QbGet(internalUserIncome.GetUserIncome{
		UserID: userIncome,
		Label:  req.Msg.GetIncome().GetByLabel(),
		Status: pbCommon.Status_STATUS_ACTIVE,
	})
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
			Response: &pbUserIDs.UpdateIncomeResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _incomePackage,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}

	defer rows.Close()

	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
			Response: &pbUserIDs.UpdateIncomeResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _incomePackage,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	userIncomeRes, err := s.userIncomeRepo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _incomePackage)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
			Response: &pbUserIDs.UpdateIncomeResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _incomePackage,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}

	qb, err = s.userIncomeRepo.QbUpdate(
		userIncome, &pbIncomes.SetLabeledIncome{
			Label:  req.Msg.GetIncome().GetAddress().GetLabel(),
			Status: req.Msg.GetIncome().GetStatus(),
			Response: &pbIncomes.SetLabeledIncome_Id{
				Id: userIncomeRes.GetIncome().GetId(),
			},
		})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
			Response: &pbUserIDs.UpdateIncomeResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _incomePackage,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}
	sqlstr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_qb, _err := s.incomeSS.Repo.QbUpdate(&pbIncomes.UpdateRequest{
			Id: userIncomeRes.GetIncome().GetId(),
			// Select: &pbIncomes.UpdateRequest_Other{},
			Status: req.Msg.GetIncome().GetAddress().GetStatus().Enum(),
		})
		if _err != nil {
			log.Error().Err(err)
			return _err
		}
		log.Info().Msg("Executing SQL '" + sqlstr + "'")
		_, upsertUserIncomeErr := common.TxWrite[pbIncomes.LabeledIncome](
			ctx, tx, sqlstr, args, s.userIncomeRepo.ScanRow)
		if upsertUserIncomeErr != nil {
			return upsertUserIncomeErr
		}

		_sqlstr, _args, _ := _qb.GenerateSQL()
		log.Info().Msg("Executing SQL '" + _sqlstr + "'")
		_, updateIncomeIncomeErr := common.TxWrite[pbIncomes.Income](
			ctx, tx, _sqlstr, _args, s.incomeSS.Repo.ScanRow)
		if updateIncomeIncomeErr != nil {
			return updateIncomeIncomeErr
		}
		return nil
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"update", _incomePackage, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
			Response: &pbUserIDs.UpdateIncomeResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _incomePackage,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}
	return connect.NewResponse(&pbUserIDs.UpdateIncomeResponse{
		Response: &pbUserIDs.UpdateIncomeResponse_Income{
			Income: &pbUserIDs.Income{
				User:   req.Msg.GetUser(),
				Income: userIncomeRes,
			},
		},
	}), nil
}
