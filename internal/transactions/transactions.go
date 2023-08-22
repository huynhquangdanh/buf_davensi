package transactions

import (
	"context"
	"database/sql"
	"fmt"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbUoMs "davensi.com/core/gen/uoms"
	pbUoMsConnect "davensi.com/core/gen/uoms/uomsconnect"
)

const (
	_package          = "transactions"
	_tableName        = "core.transactions"
	_entityName       = "Transaction"
	_entityNamePlural = "Transactions"
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	pbUoMsConnect.UnimplementedServiceHandler
	db *pgx.Conn
}

func NewServiceServer(db *pgx.Conn) *ServiceServer {
	return &ServiceServer{
		db: db,
	}
}

func (s *ServiceServer) Create(ctx context.Context, req *connect.Request[pbUoMs.CreateRequest]) (*connect.Response[pbUoMs.CreateResponse], error) { //nolint:lll
	// Verify that Type and Symbol are specified
	if req.Msg.Type == 0 || req.Msg.Symbol == "" {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", "type and symbol must be specified")
		log.Error().Err(_err)
		return connect.NewResponse(&pbUoMs.CreateResponse{
			Response: &pbUoMs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Verify that Type/Symbol is unique
	uom, err := s.Get(ctx, &connect.Request[pbUoMs.GetRequest]{
		Msg: &pbUoMs.GetRequest{
			Select: &pbUoMs.Select{
				Select: &pbUoMs.Select_ByTypeSymbol{
					ByTypeSymbol: &pbUoMs.TypeSymbol{
						Type:   req.Msg.GetType(),
						Symbol: req.Msg.GetSymbol(),
					},
				},
			},
		},
	})
	if uom.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"create", _entityName, "type/symbol = '"+req.Msg.GetType().Enum().String()+"/"+req.Msg.GetSymbol()+"'", "Type/Symbol")
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUoMs.CreateResponse{
			Response: &pbUoMs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), err
	}
	// Verify that Type/Symbol does not exist (actually expecting ERROR_CODE_NOT_FOUND)
	if uom.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return connect.NewResponse(&pbUoMs.CreateResponse{
			Response: &pbUoMs.CreateResponse_Error{
				Error: uom.Msg.GetError(),
			},
		}), err
	}

	sqlstr := "INSERT INTO " + _tableName + "(type, symbol, "
	sqlv := fmt.Sprintf(" VALUES(%d, '%s', ", req.Msg.GetType(), req.Msg.GetSymbol())

	// Populate fields and values according to what's specified in the creation request message
	if req.Msg.Name != nil {
		sqlstr += "name, " //nolint:goconst
		sqlv += "'" + req.Msg.GetName() + "', "
	}
	if req.Msg.Icon != nil {
		sqlstr += "icon, " //nolint:goconst
		sqlv += "'" + req.Msg.GetIcon() + "', "
	}
	if req.Msg.ManagedDecimals != nil {
		sqlstr += "managed_decimals, " //nolint:goconst
		sqlv += fmt.Sprintf("%d, ", req.Msg.GetManagedDecimals())
	}
	if req.Msg.DisplayedDecimals != nil {
		sqlstr += "displayed_decimals, " //nolint:goconst
		sqlv += fmt.Sprintf("%d, ", req.Msg.GetDisplayedDecimals())
	}
	if req.Msg.ReportingUnit != nil {
		sqlstr += "reporting_unit, " //nolint:goconst
		sqlv += fmt.Sprintf("%t, ", req.Msg.GetReportingUnit())
	}
	if req.Msg.Status != nil {
		sqlstr += "status, "
		sqlv += fmt.Sprintf("%d, ", uint32(req.Msg.GetStatus().Number()))
	}
	sqlstr = sqlstr[:len(sqlstr)-2] + ")"
	sqlv = sqlv[:len(sqlv)-2] + ");"
	sqlstr += sqlv

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlstr)
		return err
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
			"type/symbol = '"+req.Msg.GetType().Enum().String()+"/"+req.Msg.GetSymbol()+"'",
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUoMs.CreateResponse{
			Response: &pbUoMs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	newUoM, err := s.Get(ctx, &connect.Request[pbUoMs.GetRequest]{
		Msg: &pbUoMs.GetRequest{
			Select: &pbUoMs.Select{
				Select: &pbUoMs.Select_ByTypeSymbol{
					ByTypeSymbol: &pbUoMs.TypeSymbol{
						Type:   req.Msg.GetType(),
						Symbol: req.Msg.GetSymbol(),
					},
				},
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to create %s with type/symbol = '%s'",
			_entityName,
			req.Msg.GetType().Enum().String()+"/"+req.Msg.GetSymbol(),
		)
		return connect.NewResponse(&pbUoMs.CreateResponse{
			Response: &pbUoMs.CreateResponse_Error{
				Error: newUoM.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s with type/symbol = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetType().Enum().String()+"/"+req.Msg.GetSymbol(), newUoM.Msg.GetUom().GetId())
	return connect.NewResponse(&pbUoMs.CreateResponse{
		Response: &pbUoMs.CreateResponse_Uom{
			Uom: newUoM.Msg.GetUom(),
		},
	}), nil
}

func (s *ServiceServer) Update(ctx context.Context, req *connect.Request[pbUoMs.UpdateRequest]) (*connect.Response[pbUoMs.UpdateResponse], error) { //nolint:lll,funlen,gocyclo
	var id string
	var old_type uint32   //nolint:stylecheck
	var old_symbol string //nolint:stylecheck
	var sel string
	switch req.Msg.Select.GetSelect().(type) {
	case *pbUoMs.Select_ById:
		// Verify that Id is specified
		if req.Msg.Select.GetById() == "" {
			_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
				"updating '"+_entityName+"'", "id must be specified")
			log.Error().Err(_err)
			return connect.NewResponse(&pbUoMs.UpdateResponse{
				Response: &pbUoMs.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			}), _err
		}
		id = req.Msg.Select.GetById()
		old_type = uint32(req.Msg.GetType().Number())
		old_symbol = req.Msg.GetSymbol()
		sel = fmt.Sprintf("id = '%s'", req.Msg.Select.GetById())
	case *pbUoMs.Select_ByTypeSymbol:
		// Verify that Type and Symbol are specified
		if req.Msg.Select.GetByTypeSymbol().Type == 0 || req.Msg.Select.GetByTypeSymbol().Symbol == "" {
			_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
				"updating '"+_entityName+"'", "type and symbol must be specified")
			log.Error().Err(_err)
			return connect.NewResponse(&pbUoMs.UpdateResponse{
				Response: &pbUoMs.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			}), _err
		}
		uom, err := s.Get(ctx, &connect.Request[pbUoMs.GetRequest]{
			Msg: &pbUoMs.GetRequest{
				Select: &pbUoMs.Select{
					Select: &pbUoMs.Select_ByTypeSymbol{
						ByTypeSymbol: &pbUoMs.TypeSymbol{
							Type:   req.Msg.Select.GetByTypeSymbol().Type,
							Symbol: req.Msg.Select.GetByTypeSymbol().Symbol,
						},
					},
				},
			},
		})
		if err != nil {
			log.Error().Err(err)
			return connect.NewResponse(&pbUoMs.UpdateResponse{
				Response: &pbUoMs.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    uom.Msg.GetError().Code,
						Package: _package,
						Text:    "update failed: " + uom.Msg.GetError().Text,
					},
				},
			}), err
		}
		id = uom.Msg.GetUom().GetId()
		old_type = uint32(uom.Msg.GetUom().GetType().Number())
		old_symbol = uom.Msg.GetUom().GetSymbol()
		sel = fmt.Sprintf("type = %d AND symbol = '%s'", req.Msg.Select.GetByTypeSymbol().GetType(), req.Msg.Select.GetByTypeSymbol().GetSymbol())
	}

	sqlstr := "UPDATE " + _tableName + " SET "
	anyof := false // Will stay false if no field is selected for update
	if req.Msg.Type != nil && uint32(req.Msg.GetType().Number()) != old_type {
		sqlstr += fmt.Sprintf("type = %d, ", req.Msg.GetType())
		anyof = true
	}
	if req.Msg.Symbol != nil && req.Msg.GetSymbol() != old_symbol {
		sqlstr += fmt.Sprintf("symbol = '%s', ", req.Msg.GetSymbol())
		anyof = true
	}
	// Verify that the updated Type/Symbol does not exist
	if anyof {
		tt := old_type
		ts := old_symbol
		if req.Msg.Type != nil {
			tt = uint32(req.Msg.GetType().Number())
		}
		if req.Msg.Symbol != nil {
			ts = req.Msg.GetSymbol()
		}
		uom, err := s.Get(ctx, &connect.Request[pbUoMs.GetRequest]{
			Msg: &pbUoMs.GetRequest{
				Select: &pbUoMs.Select{
					Select: &pbUoMs.Select_ByTypeSymbol{
						ByTypeSymbol: &pbUoMs.TypeSymbol{
							Type:   pbUoMs.Type(tt),
							Symbol: ts,
						},
					},
				},
			},
		})
		if err == nil || uom.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
				"update", _entityName, "type/symbol = '"+pbUoMs.Type(tt).Enum().String()+"/"+ts+"'", "Type/Symbol")
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbUoMs.UpdateResponse{
				Response: &pbUoMs.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			}), err
		}
		// Verify that Type/Symbol does not exist (actually expecting ERROR_CODE_NOT_FOUND)
		if uom.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
			return connect.NewResponse(&pbUoMs.UpdateResponse{
				Response: &pbUoMs.UpdateResponse_Error{
					Error: uom.Msg.GetError(),
				},
			}), err
		}
	}
	if req.Msg.Name != nil {
		sqlstr += fmt.Sprintf("name = '%s', ", req.Msg.GetName())
		anyof = true
	}
	if req.Msg.Icon != nil {
		sqlstr += fmt.Sprintf("icon = '%s', ", req.Msg.GetIcon())
		anyof = true
	}
	if req.Msg.ManagedDecimals != nil {
		sqlstr += fmt.Sprintf("managed_decimals = %d, ", req.Msg.GetManagedDecimals())
		anyof = true
	}
	if req.Msg.DisplayedDecimals != nil {
		sqlstr += fmt.Sprintf("displayed_decimals = %d, ", req.Msg.GetDisplayedDecimals())
		anyof = true
	}
	if req.Msg.ReportingUnit != nil {
		sqlstr += fmt.Sprintf("reporting_unit = %t, ", req.Msg.GetReportingUnit())
		anyof = true
	}
	if req.Msg.Status != nil {
		sqlstr += fmt.Sprintf("status = %d, ", req.Msg.GetStatus())
		anyof = true
	}
	// Execute only if at least one field is speficied for update
	if anyof {
		sqlstr = sqlstr[:len(sqlstr)-2] + " WHERE " + sel + ";"
		log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
		if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, sqlstr)
			return err
		}); err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
				"updating",
				_entityName,
				sel,
			)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbUoMs.UpdateResponse{
				Response: &pbUoMs.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error() + "(" + err.Error() + ")",
					},
				},
			}), err
		}
	}

	// Start building the response from here
	updatedUoM, err := s.Get(ctx, &connect.Request[pbUoMs.GetRequest]{
		Msg: &pbUoMs.GetRequest{
			Select: &pbUoMs.Select{
				Select: &pbUoMs.Select_ById{
					ById: id,
				},
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to update %s with %s", _entityName, sel)
		return connect.NewResponse(&pbUoMs.UpdateResponse{
			Response: &pbUoMs.UpdateResponse_Error{
				Error: updatedUoM.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbUoMs.UpdateResponse{
		Response: &pbUoMs.UpdateResponse_Uom{
			Uom: updatedUoM.Msg.GetUom(),
		},
	}), nil
}

func (s *ServiceServer) Get(ctx context.Context, req *connect.Request[pbUoMs.GetRequest]) (*connect.Response[pbUoMs.GetResponse], error) { //nolint:lll,funlen
	sqlstr := "SELECT "
	sqlstr += "id, "
	sqlstr += "type, "
	sqlstr += "symbol, "
	sqlstr += "name, "
	sqlstr += "icon, "
	sqlstr += "managed_decimals, "
	sqlstr += "displayed_decimals, "
	sqlstr += "reporting_unit, "
	sqlstr += "status"
	sqlstr += " FROM " + _tableName + " WHERE "
	var sel string
	switch req.Msg.GetSelect().Select.(type) {
	case *pbUoMs.Select_ById:
		sel = fmt.Sprintf("id = '%s'", req.Msg.GetSelect().GetById())
	case *pbUoMs.Select_ByTypeSymbol:
		if req.Msg.GetSelect().Select != nil {
			sel = fmt.Sprintf(
				"type = %d AND symbol = '%s'",
				req.Msg.GetSelect().GetByTypeSymbol().Type,
				req.Msg.GetSelect().GetByTypeSymbol().GetSymbol())
		}
	}
	sqlstr += sel + ";"
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	rows, err := s.db.Query(ctx, sqlstr)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
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
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	var id string
	var uomType uint32
	var symbol string
	// Nullable field must use sql.* type or a scan error will be thrown
	var nullableName sql.NullString
	var nullableIcon sql.NullString
	var managedDecimals uint32
	var displayedDecimals uint32
	var reportingUnit bool
	var status uint32
	if err := rows.Scan(
		&id,
		&uomType,
		&symbol,
		&nullableName,
		&nullableIcon,
		&managedDecimals,
		&displayedDecimals,
		&reportingUnit,
		&status,
	); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	if rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	// Nullable field processing
	var name *string = nil
	if nullableName.Valid {
		name = &nullableName.String
	}
	var icon *string = nil
	if nullableIcon.Valid {
		icon = &nullableIcon.String
	}
	// Start building the response from here
	return connect.NewResponse(&pbUoMs.GetResponse{
		Response: &pbUoMs.GetResponse_Uom{
			Uom: &pbUoMs.UoM{
				Id:                id,
				Type:              pbUoMs.Type(uomType),
				Symbol:            symbol,
				Name:              name,
				Icon:              icon,
				ManagedDecimals:   managedDecimals,
				DisplayedDecimals: displayedDecimals,
				ReportingUnit:     reportingUnit,
				Status:            pbCommon.Status(status),
			},
		},
	}), nil
}

func (s *ServiceServer) GetList(ctx context.Context, req *connect.Request[pbUoMs.GetListRequest], res *connect.ServerStream[pbUoMs.GetListResponse]) error { //nolint:lll,funlen,gocyclo
	sqlsel := "SELECT "
	sqlsel += "id, "
	sqlsel += "type, "
	sqlsel += "symbol, "
	sqlsel += "name, "
	sqlsel += "icon, "
	sqlsel += "managed_decimals, "
	sqlsel += "displayed_decimals, "
	sqlsel += "reporting_unit, "
	sqlsel += "status"
	sqlsel += " FROM " + _tableName
	sqlstr := " WHERE "
	anyof := false // Will stay false if no field is selected
	if req.Msg.Type != nil {
		sqlstr += "("
		for _, t := range req.Msg.GetType().GetList() {
			sqlstr += fmt.Sprintf("type = %d OR ", uint32(t.Number()))
		}
		sqlstr = sqlstr[:len(sqlstr)-4] + ") AND "
		anyof = true
	}
	if req.Msg.Symbol != nil {
		sqlstr += "symbol LIKE '%" + req.Msg.GetSymbol() + "%' AND "
		anyof = true
	}
	if req.Msg.Name != nil {
		sqlstr += "symbol LIKE '%" + req.Msg.GetName() + "%' AND "
		anyof = true
	}
	if req.Msg.Icon != nil {
		sqlstr += "symbol LIKE '%" + req.Msg.GetIcon() + "%' AND "
		anyof = true
	}
	if req.Msg.ManagedDecimals != nil { //nolint:dupl
		sqlstr += "("
		for _, v := range req.Msg.GetManagedDecimals().GetList() {
			switch v.GetSelect().(type) {
			case *pbCommon.UInt32Values_Single:
				sqlstr += fmt.Sprintf("managed_decimals = %d OR ", v.GetSelect().(*pbCommon.UInt32Values_Single).Single)
			case *pbCommon.UInt32Values_Range:
				if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil && v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("(managed_decimals >= %d AND ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("(managed_decimals > %d AND ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("managed_decimals <= %d) OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("managed_decimals < %d) OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
				} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil {
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("managed_decimals >= %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("managed_decimals > %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
				} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("managed_decimals <= %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("managed_decimals < %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
				}
			}
		}
		sqlstr = sqlstr[:len(sqlstr)-4] + ") AND "
		anyof = true
	}
	if req.Msg.DisplayedDecimals != nil { //nolint:dupl
		sqlstr += "("
		for _, v := range req.Msg.GetDisplayedDecimals().GetList() {
			switch v.GetSelect().(type) {
			case *pbCommon.UInt32Values_Single:
				sqlstr += fmt.Sprintf("displayed_decimals = %d OR ", v.GetSelect().(*pbCommon.UInt32Values_Single).Single)
			case *pbCommon.UInt32Values_Range:
				if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil && v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("(displayed_decimals >= %d AND ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("(displayed_decimals > %d AND ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("displayed_decimals <= %d) OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("displayed_decimals < %d) OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
				} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil {
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("displayed_decimals >= %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("displayed_decimals > %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
				} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
					switch v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(type) {
					case *pbCommon.UInt32Boundary_Incl:
						sqlstr += fmt.Sprintf("displayed_decimals <= %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl)
					case *pbCommon.UInt32Boundary_Excl:
						sqlstr += fmt.Sprintf("displayed_decimals < %d OR ",
							v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl)
					}
				}
			}
		}
		sqlstr = sqlstr[:len(sqlstr)-4] + ") AND "
		anyof = true
	}
	if req.Msg.ReportingUnit != nil {
		sqlstr += fmt.Sprintf("reporting_unit = %t AND ", req.Msg.GetReportingUnit())
		anyof = true
	}
	if req.Msg.Status != nil {
		sqlstr += "("
		for _, s := range req.Msg.GetStatus().GetList() {
			sqlstr += fmt.Sprintf("status = %d OR ", uint32(s.Number()))
		}
		sqlstr = sqlstr[:len(sqlstr)-4] + ") AND "
		anyof = true
	}
	if anyof {
		sqlstr = sqlstr[:len(sqlstr)-5] + ";"
		sqlstr = sqlsel + sqlstr
	} else {
		sqlstr = sqlsel + ";"
	}
	log.Info().Msg("Executing SQL: " + sqlstr)
	rows, err := s.db.Query(ctx, sqlstr)
	if err != nil { //nolint:dupl
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
		log.Error().Err(err).Msg(_err.Error())

		if errSend := res.Send(&pbUoMs.GetListResponse{
			Response: &pbUoMs.GetListResponse_Error{
				Error: &pbCommon.Error{
					Code: _errno,
					Text: _err.Error() + " (" + err.Error() + ")",
				},
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
			_err = _errSend
		}
		return _err
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		var id string
		var uomtype int32
		var symbol string
		var nullable_name sql.NullString //nolint:stylecheck
		var nullable_icon sql.NullString //nolint:stylecheck
		var managedDecimals uint32
		var displayedDecimals uint32
		var reporting_unit bool //nolint:stylecheck
		var status uint32
		if err := rows.Scan(
			&id,
			&uomtype,
			&symbol,
			&nullable_name,
			&nullable_icon,
			&managedDecimals,
			&displayedDecimals,
			&reporting_unit,
			&status,
		); err != nil { //nolint:dupl
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(err).Msg(_err.Error())

			if errSend := res.Send(&pbUoMs.GetListResponse{
				Response: &pbUoMs.GetListResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error() + " (" + err.Error() + ")",
					},
				},
			}); errSend != nil {
				_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
				_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
				log.Error().Err(errSend).Msg(_errSend.Error())
				_err = _errSend
			}
			return _err
		}

		// Nullable field processing
		var name *string = nil
		if nullable_name.Valid {
			name = &nullable_name.String
		}
		var icon *string = nil
		if nullable_icon.Valid {
			icon = &nullable_icon.String
		}

		if errSend := res.Send(&pbUoMs.GetListResponse{
			Response: &pbUoMs.GetListResponse_Uom{
				Uom: &pbUoMs.UoM{
					Id:                id,
					Type:              pbUoMs.Type(uomtype),
					Symbol:            symbol,
					Name:              name,
					Icon:              icon,
					ManagedDecimals:   managedDecimals,
					DisplayedDecimals: displayedDecimals,
					ReportingUnit:     reporting_unit,
					Status:            pbCommon.Status(status),
				},
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}
