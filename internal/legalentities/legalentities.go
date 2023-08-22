package legalentities

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbAddresses "davensi.com/core/gen/addresses"
	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbCountries "davensi.com/core/gen/countries"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbLegalEntitiesConnect "davensi.com/core/gen/legalentities/legalentitiesconnect"
	pbUoms "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/addresses"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/contacts"
	"davensi.com/core/internal/countries"
	legalentitiesaddresses "davensi.com/core/internal/legalentities_addresses"
	"davensi.com/core/internal/uoms"
)

const (
	_package          = "legalentities"
	_entityName       = "Legal Entity"
	_entityNamePlural = "Legal Entities"
)

// For singleton UoMs export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the LegalEntitiesService API
type ServiceServer struct {
	Repo LegalEntitiesRepository
	pbLegalEntitiesConnect.UnimplementedServiceHandler
	db                     *pgxpool.Pool
	countriesSS            *countries.ServiceServer
	uomsSS                 *uoms.ServiceServer
	addressesSS            *addresses.ServiceServer
	contactsSS             *contacts.ServiceServer
	addressRepo            addresses.AddressRepository
	contactRepo            contacts.ContactRepository
	legalEntityAddressRepo legalentitiesaddresses.LegalEntityAddressesRepository
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:                   *NewLegalEntitiesRepository(db),
		db:                     db,
		countriesSS:            countries.GetSingletonServiceServer(db),
		uomsSS:                 uoms.GetSingletonServiceServer(db),
		addressesSS:            addresses.GetSingletonServiceServer(db),
		contactsSS:             contacts.GetSingletonServiceServer(db),
		addressRepo:            addresses.GetSingletonServiceServer(db).Repo,
		legalEntityAddressRepo: *legalentitiesaddresses.GetSingletonRepository(db),
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.CreateRequest],
) (*connect.Response[pbLegalEntities.CreateResponse], error) {
	if errCreation := s.validateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbLegalEntities.CreateResponse{
			Response: &pbLegalEntities.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	qb, errCreateQb := s.Repo.QbInsert(req.Msg)
	if errCreateQb != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"create",
			_package,
			errCreateQb.Error(),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbLegalEntities.CreateResponse{
			Response: &pbLegalEntities.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	newLegalEntity, errExecInsert := common.ExecuteTxWrite[pbLegalEntities.LegalEntity](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanMainEntity,
	)
	if errExecInsert != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			fmt.Sprintf(
				"type/name = '%s/%s' with err: %s",
				req.Msg.GetType().Enum().String(),
				req.Msg.GetName(),
				errExecInsert.Error(),
			),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbLegalEntities.CreateResponse{
			Response: &pbLegalEntities.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s with name = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetName(), newLegalEntity.Id)
	return connect.NewResponse(&pbLegalEntities.CreateResponse{
		Response: &pbLegalEntities.CreateResponse_LegalEntity{
			LegalEntity: newLegalEntity,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.UpdateRequest],
) (*connect.Response[pbLegalEntities.UpdateResponse], error) {
	commonErr := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"updating",
		_entityName,
		"",
	)

	if errQueryUpdate := s.validateUpdateQuery(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbLegalEntities.UpdateResponse{
			Response: &pbLegalEntities.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getLegalEntityResponse, errGetOldLegalEntity := s.GetOneMainEntity(context.Background(), &connect.Request[pbLegalEntities.GetRequest]{
		Msg: &pbLegalEntities.GetRequest{
			Select: req.Msg.Select,
		},
	})

	if errGetOldLegalEntity != nil {
		log.Error().Err(errGetOldLegalEntity)
		return connect.NewResponse(&pbLegalEntities.UpdateResponse{
			Response: &pbLegalEntities.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getLegalEntityResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getLegalEntityResponse.Msg.GetError().Text,
				},
			},
		}), errGetOldLegalEntity
	}

	legalEntityBeforeUpdate := getLegalEntityResponse.Msg.GetLegalEntity()

	if errValidateMsgUpdate := s.validateUpdateValue(req.Msg, legalEntityBeforeUpdate); errValidateMsgUpdate != nil {
		log.Error().Err(errValidateMsgUpdate.Err)
		return connect.NewResponse(&pbLegalEntities.UpdateResponse{
			Response: &pbLegalEntities.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errValidateMsgUpdate.Code,
					Package: _package,
					Text:    errValidateMsgUpdate.Err.Error(),
				},
			},
		}), errValidateMsgUpdate.Err
	}

	qb, errGenUpdate := s.Repo.QbUpdate(req.Msg)
	if errGenUpdate != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errGenUpdate.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.UpdateResponse{
			Response: &pbLegalEntities.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL '" + sqlstr + "'")
	updatedLegalEntity, errExcute := common.ExecuteTxWrite[pbLegalEntities.LegalEntity](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanMainEntity,
	)
	if errExcute != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errExcute.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.UpdateResponse{
			Response: &pbLegalEntities.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbLegalEntities.UpdateResponse{
		Response: &pbLegalEntities.UpdateResponse_LegalEntity{
			LegalEntity: updatedLegalEntity,
		},
	}), nil
}

func (s *ServiceServer) GetOneMainEntity(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.GetRequest],
) (*connect.Response[pbLegalEntities.GetResponse], error) {
	commonErr := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"fetching",
		_entityName,
		"",
	)

	if errQueryGet := ValidateSelect(req.Msg.GetSelect(), "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, errQueryRows := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errQueryRows != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errQueryRows.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	defer rows.Close()

	if !rows.Next() {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND).
			UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}
	legalEntity, errScanRow := s.Repo.ScanMainEntity(rows)
	if errScanRow != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).
			UpdateMessage(errScanRow.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}
	if rows.Next() {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND).
			UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbLegalEntities.GetResponse{
		Response: &pbLegalEntities.GetResponse_LegalEntity{
			LegalEntity: legalEntity,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.GetRequest],
) (*connect.Response[pbLegalEntities.GetResponse], error) {
	commonErr := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, "fetching", _entityName, "")

	if errQueryGet := ValidateSelect(req.Msg.GetSelect(), "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	var (
		qbIncorporationCountry = s.countriesSS.Repo.QbGetList(&pbCountries.GetListRequest{})
		qbUom1                 = s.uomsSS.Repo.QbGetList(&pbUoms.GetListRequest{})
		qbUom2                 = s.uomsSS.Repo.QbGetList(&pbUoms.GetListRequest{})
		qbUom3                 = s.uomsSS.Repo.QbGetList(&pbUoms.GetListRequest{})
	)

	var (
		filterICStr, filterICArgs     = qbIncorporationCountry.Filters.GenerateSQL() // IC = Incorporation Country
		filterUom1Str, filterUom1Args = qbUom1.Filters.GenerateSQL()
		filterUom2Str, filterUom2Args = qbUom2.Filters.GenerateSQL()
		filterUom3Str, filterUom3Args = qbUom3.Filters.GenerateSQL()
	)

	qb := s.Repo.QbGetOne(req.Msg)
	qb.Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.incorporation_country_id = countries.id", qbIncorporationCountry.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.currency1_id = uoms.id", qbUom1.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.currency2_id = uoms.id", qbUom2.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.currency3_id = uoms.id", qbUom3.TableName)).
		Select(strings.Join(qbIncorporationCountry.SelectFields, ", ")).
		Select(strings.Join(qbUom1.SelectFields, ", ")).
		Select(strings.Join(qbUom2.SelectFields, ", ")).
		Select(strings.Join(qbUom3.SelectFields, ", ")).
		Where(filterICStr, filterICArgs...).
		Where(filterUom1Str, filterUom1Args...).
		Where(filterUom2Str, filterUom2Args...).
		Where(filterUom3Str, filterUom3Args...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	// Transform sqlstr to give aliases to core.uoms because there are multiple left-joins on that table
	replaceCount := 9
	sqlstr = strings.Replace(sqlstr, "uoms", "currency1", replaceCount)
	sqlstr = strings.Replace(sqlstr, "uoms", "currency2", replaceCount)
	sqlstr = strings.Replace(sqlstr, "uoms", "currency3", replaceCount)
	sqlstr = strings.Replace(sqlstr, "core.uoms ON legalentities.currency1_id = uoms.id",
		"core.uoms currency1 ON legalentities.currency1_id = currency1.id", 1)
	sqlstr = strings.Replace(sqlstr, "core.uoms ON legalentities.currency2_id = uoms.id",
		"core.uoms currency2 ON legalentities.currency2_id = currency2.id", 1)
	sqlstr = strings.Replace(sqlstr, "core.uoms ON legalentities.currency3_id = uoms.id",
		"core.uoms currency3 ON legalentities.currency3_id = currency3.id", 1)

	log.Info().Msg("Executing GET SQL '" + sqlstr + "'")
	rows, errQueryRows := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errQueryRows != nil {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).UpdateMessage(errQueryRows.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	defer rows.Close()

	if !rows.Next() {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND).UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	legalEntity, errScanRow := s.Repo.ScanWithRelationship(rows)
	if errScanRow != nil {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).UpdateMessage(errScanRow.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	if rows.Next() {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND).UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbLegalEntities.GetResponse{
			Response: &pbLegalEntities.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	// add address list if needed
	var err error
	if req.Msg.WithAddresses != nil && *req.Msg.WithAddresses {
		legalEntity, err = s.GetAddressList(ctx, legalEntity.Id, legalEntity)
		if err != nil {
			commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).UpdateMessage("cannot find related address list")
			log.Error().Err(commonErr.Err)
			return connect.NewResponse(&pbLegalEntities.GetResponse{
				Response: &pbLegalEntities.GetResponse_Error{
					Error: &pbCommon.Error{
						Code:    commonErr.Code,
						Package: _package,
						Text:    commonErr.Err.Error(),
					},
				},
			}), commonErr.Err
		}
	}

	// add contact list if needed
	if req.Msg.WithContacts != nil && *req.Msg.WithContacts {
		legalEntity, err = s.GetContactList(ctx, legalEntity.Id, legalEntity)
		if err != nil {
			commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).UpdateMessage("cannot find related contact list")
			log.Error().Err(commonErr.Err)
			return connect.NewResponse(&pbLegalEntities.GetResponse{
				Response: &pbLegalEntities.GetResponse_Error{
					Error: &pbCommon.Error{
						Code:    commonErr.Code,
						Package: _package,
						Text:    commonErr.Err.Error(),
					},
				},
			}), commonErr.Err
		}
	}

	// Start building the response from here
	return connect.NewResponse(&pbLegalEntities.GetResponse{
		Response: &pbLegalEntities.GetResponse_LegalEntity{
			LegalEntity: legalEntity,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.GetListRequest],
	res *connect.ServerStream[pbLegalEntities.GetListResponse],
) error {
	if req.Msg.IncorporationCountry == nil {
		req.Msg.IncorporationCountry = &pbCountries.GetListRequest{}
	}

	if req.Msg.Currency1 == nil {
		req.Msg.Currency1 = &pbUoms.GetListRequest{}
	}

	if req.Msg.Currency2 == nil {
		req.Msg.Currency2 = &pbUoms.GetListRequest{}
	}

	if req.Msg.Currency3 == nil {
		req.Msg.Currency3 = &pbUoms.GetListRequest{}
	}

	var (
		qbIC   = s.countriesSS.Repo.QbGetList(req.Msg.IncorporationCountry) // IC = Incorporation Country
		qbUom1 = s.uomsSS.Repo.QbGetList(req.Msg.Currency1)
		qbUom2 = s.uomsSS.Repo.QbGetList(req.Msg.Currency2)
		qbUom3 = s.uomsSS.Repo.QbGetList(req.Msg.Currency3)

		filterICStr, filterICArgs     = qbIC.Filters.GenerateSQL()
		filterUom1Str, filterUom1Args = qbUom1.Filters.GenerateSQL()
		filterUom2Str, filterUom2Args = qbUom2.Filters.GenerateSQL()
		filterUom3Str, filterUom3Args = qbUom3.Filters.GenerateSQL()

		// addressIDs string // in case with_addresses == TRUE
		contactIDs string // in case with_contacts == TRUE
	)

	// Assigning different aliases to core.uoms, because there are multiple JOINs on that same table
	filterUom1Str = strings.ReplaceAll(filterUom1Str, "uoms", "currency1")
	filterUom2Str = strings.ReplaceAll(filterUom2Str, "uoms", "currency2")
	filterUom3Str = strings.ReplaceAll(filterUom3Str, "uoms", "currency3")

	qb := s.Repo.QbGetList(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.incorporation_country_id = countries.id", qbIC.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.currency1_id = uoms.id", qbUom1.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.currency2_id = uoms.id", qbUom2.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.currency3_id = uoms.id", qbUom3.TableName)).
		Select(strings.Join(qbIC.SelectFields, ", ")).
		Select(strings.Join(qbUom1.SelectFields, ", ")).
		Select(strings.Join(qbUom2.SelectFields, ", ")).
		Select(strings.Join(qbUom3.SelectFields, ", ")).
		Where(filterICStr, filterICArgs...).
		Where(filterUom1Str, filterUom1Args...).
		Where(filterUom2Str, filterUom2Args...).
		Where(filterUom3Str, filterUom3Args...)

	if req.Msg.Addresses != nil {
		var (
			qbAddress                           = s.addressesSS.Repo.QbGetList(req.Msg.Addresses)
			filterAddressStr, filterAddressArgs = qbAddress.Filters.GenerateSQL()
		)
		qb.Join(fmt.Sprintf("LEFT JOIN %s ON legalentities.id IN (SELECT legalentity_id FROM core.legalentities_addresses)",
			qbAddress.TableName)).
			Where(filterAddressStr, filterAddressArgs...)
	}

	if req.Msg.Contacts != nil {
		contactList, err := s.contactsSS.GetListInternal(ctx, req.Msg.GetContacts())
		if err != nil {
			log.Error().Err(err)
			return err
		}

		for _, v := range contactList {
			contactIDs = contactIDs + "'" + v.Id + "', "
		}
		contactIDs = strings.TrimSuffix(contactIDs, ", ")

		qb.Join(fmt.Sprintf("JOIN (SELECT legalentity_id FROM core.legalentities_contacts "+
			"WHERE contact_id IN (%s)) s ON legalentities.id = s.legalentity_id", contactIDs))
	}

	sqlstr, sqlArgs := generateAndTransformSQL(qb)
	log.Info().Msg("Executing SQL: " + sqlstr)
	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbLegalEntities.GetListResponse{
					Response: &pbLegalEntities.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}
	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		legalEntity, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbLegalEntities.GetListResponse{
						Response: &pbLegalEntities.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		legalEntity, err = s.appendAddressesContacts(ctx, req.Msg, legalEntity)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbLegalEntities.GetListResponse{
						Response: &pbLegalEntities.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbLegalEntities.GetListResponse{
			Response: &pbLegalEntities.GetListResponse_LegalEntity{
				LegalEntity: legalEntity,
			},
		}); errSend != nil {
			_errSend := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR)], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.DeleteRequest],
) (*connect.Response[pbLegalEntities.DeleteResponse], error) {
	deletedLegalEntity, err := s.Update(ctx, connect.NewRequest[pbLegalEntities.UpdateRequest](&pbLegalEntities.UpdateRequest{
		Select: req.Msg.Select,
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbLegalEntities.DeleteResponse{
			Response: &pbLegalEntities.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedLegalEntity.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedLegalEntity.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbLegalEntities.DeleteResponse{
		Response: &pbLegalEntities.DeleteResponse_LegalEntity{
			LegalEntity: deletedLegalEntity.Msg.GetLegalEntity(),
		},
	}), nil
}

func (s *ServiceServer) queryLegalEntityID(ctx context.Context, req *pbLegalEntities.Select) (pgx.Rows, error) {
	var (
		row pgx.Rows
		err error
	)
	switch req.GetSelect().(type) {
	case *pbLegalEntities.Select_ByName:
		row, err = s.db.Query(ctx, "SELECT id FROM core.legalentities WHERE name = $1", req.GetByName())
	case *pbLegalEntities.Select_ById:
		row, err = s.db.Query(ctx, "SELECT id FROM core.legalentities WHERE id = $1", req.GetById())
	}
	return row, err
}

func (s *ServiceServer) GetLegalEntityID(ctx context.Context, req *pbLegalEntities.Select) (string, error) {
	var legalEntityID string
	row, err := s.queryLegalEntityID(ctx, req)
	if err != nil {
		return "", err
	}
	defer row.Close()
	if row.Next() {
		legalEntityID, err = s.Repo.ScanID(row)
		if err != nil {
			return "", err
		}
	}
	return legalEntityID, nil
}

func (s *ServiceServer) SeparateAddressQuery(req *pbLegalEntities.SetAddressesRequest) (
	updateUserAddresses []*pbAddresses.SetLabeledAddress,
	insertAddresses []*pbAddresses.SetLabeledAddress,
) {
	setLabelAddresses := req.Addresses.GetList()
	for _, labelAddress := range setLabelAddresses {
		if labelAddress.GetId() != "" {
			// id is not null
			insertAddresses = append(insertAddresses, labelAddress)
		} else if labelAddress.GetAddress() != nil {
			updateUserAddresses = append(updateUserAddresses, labelAddress)
		}
	}
	return updateUserAddresses, insertAddresses
}

func (s *ServiceServer) SeparateContactsQuery(req *pbLegalEntities.SetContactsRequest) (
	updateContacts []*pbContacts.SetLabeledContact,
	insertContacts []*pbContacts.SetLabeledContact,
) {
	setLabelContacts := req.GetContacts().GetList()
	for _, labelContact := range setLabelContacts {
		if labelContact.GetId() != "" {
			// id is not null
			insertContacts = append(insertContacts, labelContact)
		} else if labelContact.GetContact() != nil {
			updateContacts = append(updateContacts, labelContact)
		}
	}
	return updateContacts, insertContacts
}

func (s *ServiceServer) SetAddresses(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.SetAddressesRequest],
) (*connect.Response[pbLegalEntities.SetAddressesResponse], error) {
	if errCode, err := s.ValidateSetAddresses(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
			Response: &pbLegalEntities.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
			Response: &pbLegalEntities.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if _err := tx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()

	legalEntityID, err := s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
			Response: &pbLegalEntities.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	// get query builder for upsert user_addresses
	// separate update query and insert query with address
	updateUserAddresses, insertAddresses := s.SeparateAddressQuery(req.Msg)
	// insert address first and get id
	qbInsertAddress, err := s.addressRepo.QbInsertManyAddress(req.Msg.GetAddresses())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
			Response: &pbLegalEntities.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	sqlInsertAddress, sqlInsertAddressArgs, _ := qbInsertAddress.GenerateSQL()
	log.Info().Msgf("query insert address %s with args %s", sqlInsertAddress, sqlInsertAddressArgs)

	rows, err := tx.Query(ctx, sqlInsertAddress, sqlInsertAddressArgs...)
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
			Response: &pbLegalEntities.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	defer rows.Close()
	countInsert := 0
	res := map[string]*pbAddresses.SetLabeledAddress{} // {<address_id>: {SetLabelAddress}}
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
				Response: &pbLegalEntities.SetAddressesResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
		res[id] = insertAddresses[countInsert]
		countInsert++
	}

	// query update
	for _, address := range updateUserAddresses {
		res[address.GetId()] = address
	}
	upsertLabelAddresses := []*pbAddresses.SetLabeledAddress{}
	for k, v := range res {
		labelAddress := &pbAddresses.SetLabeledAddress{
			Label: v.Label,
			Select: &pbAddresses.SetLabeledAddress_Id{
				Id: k,
			},
			Status:          v.Status,
			MainAddress:     v.MainAddress,
			OwnershipStatus: v.OwnershipStatus,
		}
		upsertLabelAddresses = append(upsertLabelAddresses, labelAddress)
	}

	result, err := s.UpsertAddresses(ctx, tx, legalEntityID, upsertLabelAddresses)
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
			Response: &pbLegalEntities.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	if err = tx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
			Response: &pbLegalEntities.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.SetAddressesResponse{
		Response: &pbLegalEntities.SetAddressesResponse_Addresses{
			Addresses: &pbLegalEntities.AddressList{
				Id:   legalEntityID,
				Name: "",
				Addresses: &pbAddresses.LabeledAddressList{
					List: result,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) UpsertAddresses(
	ctx context.Context,
	tx pgx.Tx,
	legalEntityID string,
	upsertLabelAddresses []*pbAddresses.SetLabeledAddress,
) ([]*pbAddresses.LabeledAddress, error) {
	qbUserAddress, err := s.legalEntityAddressRepo.QbUpsertUserAddresses(legalEntityID, upsertLabelAddresses, "upsert")
	if err != nil {
		return nil, err
	}
	sqlUpsertUserAddress, sqlUpsertUserAddressArgs, _ := qbUserAddress.GenerateSQL()
	log.Info().Msgf("query upsert user address %s with args: %s", sqlUpsertUserAddress, sqlUpsertUserAddressArgs)
	rowsUpsertUserAddress, err := tx.Query(ctx, sqlUpsertUserAddress, sqlUpsertUserAddressArgs...)
	if err != nil {
		return nil, err
	}
	defer rowsUpsertUserAddress.Close()
	var result []*pbAddresses.LabeledAddress
	for rowsUpsertUserAddress.Next() {
		labelAddress, err := s.legalEntityAddressRepo.ScanRow(rowsUpsertUserAddress)
		if err != nil {
			return nil, err
		}
		result = append(result, labelAddress)
	}
	return result, nil
}

func (s *ServiceServer) AddAddresses(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.AddAddressesRequest],
) (*connect.Response[pbLegalEntities.AddAddressesResponse], error) {
	if errCode, err := s.ValidateAddAddresses(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	// get user id from req.user
	legalEntityID, err := s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	qbInsertAddress, err := s.addressRepo.QbInsertManyAddress(req.Msg.GetAddresses())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	sqlInsertAddress, sqlInsertAddressArgs, _ := qbInsertAddress.GenerateSQL()

	rows, err := tx.Query(ctx, sqlInsertAddress, sqlInsertAddressArgs...)
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	defer rows.Close()
	countInsert := 0
	res := map[string]*pbAddresses.SetLabeledAddress{} // {<address_id>: {SetLabelAddress}}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
				Response: &pbLegalEntities.AddAddressesResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
		res[id] = req.Msg.GetAddresses().GetList()[countInsert]
		countInsert++
	}

	// query update
	insertLabelAddresses := []*pbAddresses.SetLabeledAddress{}
	for k, v := range res {
		labelAddress := &pbAddresses.SetLabeledAddress{
			Label: v.Label,
			Select: &pbAddresses.SetLabeledAddress_Id{
				Id: k,
			},
			Status:          v.Status,
			MainAddress:     v.MainAddress,
			OwnershipStatus: v.OwnershipStatus,
		}
		insertLabelAddresses = append(insertLabelAddresses, labelAddress)
	}

	qbUserAddress, err := s.legalEntityAddressRepo.QbUpsertUserAddresses(legalEntityID, insertLabelAddresses, "insert")
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	sqlUpsertUserAddress, sqlUpsertUserAddressArgs, _ := qbUserAddress.GenerateSQL()
	log.Info().Msgf("query upsert user address %s with args: %s", sqlUpsertUserAddress, sqlUpsertUserAddressArgs)
	rowsUpsertUserAddress, err := tx.Query(ctx, sqlUpsertUserAddress, sqlUpsertUserAddressArgs...)
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	defer rowsUpsertUserAddress.Close()
	result := []*pbAddresses.LabeledAddress{}
	for rowsUpsertUserAddress.Next() {
		labelAddress, err := s.legalEntityAddressRepo.ScanRow(rowsUpsertUserAddress)
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
				Response: &pbLegalEntities.AddAddressesResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
		result = append(result, labelAddress)
	}

	if err = tx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
			Response: &pbLegalEntities.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.AddAddressesResponse{
		Response: &pbLegalEntities.AddAddressesResponse_Addresses{
			Addresses: &pbLegalEntities.AddressList{
				Id:   legalEntityID,
				Name: "",
				Addresses: &pbAddresses.LabeledAddressList{
					List: result,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) UpdateAddress(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.UpdateAddressRequest],
) (*connect.Response[pbLegalEntities.UpdateAddressResponse], error) {
	if errCode, err := s.ValidateUpdateAddress(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
			Response: &pbLegalEntities.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	pgxTx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
			Response: &pbLegalEntities.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if _err := pgxTx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()

	legalEntityID, err := s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
			Response: &pbLegalEntities.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	var (
		updatedAddress *pbAddresses.Address
		labelAddress   *pbAddresses.LabeledAddress
	)

	if req.Msg.GetAddress().Status != nil {
		labelAddress, err = s.UpdateLegalEntityAddressStatus(ctx, pgxTx, legalEntityID, req.Msg.GetAddress())
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
				Response: &pbLegalEntities.UpdateAddressResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
	} else {
		qbLabelAddress := s.legalEntityAddressRepo.QbGetOne(legalEntityID, req.Msg.GetAddress().ByLabel)
		sqlstr, sqlArgs, sel := qbLabelAddress.GenerateSQL()
		log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

		rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
				Response: &pbLegalEntities.UpdateAddressResponse_Error{
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
			return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
				Response: &pbLegalEntities.UpdateAddressResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			}), _err
		}
		labelAddress, err = s.legalEntityAddressRepo.ScanRow(rows)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
				Response: &pbLegalEntities.UpdateAddressResponse_Error{
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
			return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
				Response: &pbLegalEntities.UpdateAddressResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			}), _err
		}
	}

	if req.Msg.Address != nil {
		updateAddressReq := &pbAddresses.UpdateRequest{
			Id:         labelAddress.Address.Id,
			Type:       req.Msg.GetAddress().GetAddress().Type,
			Country:    req.Msg.GetAddress().GetAddress().Country,
			Building:   req.Msg.GetAddress().GetAddress().Building,
			Floor:      req.Msg.GetAddress().GetAddress().Floor,
			Unit:       req.Msg.GetAddress().GetAddress().Unit,
			StreetNum:  req.Msg.GetAddress().GetAddress().StreetNum,
			StreetName: req.Msg.GetAddress().GetAddress().StreetName,
			District:   req.Msg.GetAddress().GetAddress().District,
			Locality:   req.Msg.GetAddress().GetAddress().Locality,
			ZipCode:    req.Msg.GetAddress().GetAddress().ZipCode,
			Region:     req.Msg.GetAddress().GetAddress().Region,
			State:      req.Msg.GetAddress().GetAddress().State,
			Status:     req.Msg.GetAddress().GetAddress().Status,
		}
		qbAddress, err := s.addressRepo.QbUpdate(updateAddressReq)
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
				Response: &pbLegalEntities.UpdateAddressResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), err
		}

		sqlStr, args, sel := qbAddress.GenerateSQL()

		log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
		updatedAddress, err = common.ExecuteTxWrite[pbAddresses.Address](
			ctx,
			s.db,
			sqlStr,
			args,
			s.addressRepo.ScanRow,
		)
		if err != nil {
			errUpdate := common.CreateErrWithCode(
				pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
				"updating",
				_package,
				fmt.Sprintf("%s, %s", sel, err.Error()),
			)
			log.Error().Err(errUpdate.Err)
			return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
				Response: &pbLegalEntities.UpdateAddressResponse_Error{
					Error: &pbCommon.Error{
						Code:    errUpdate.Code,
						Package: _package,
						Text:    errUpdate.Err.Error(),
					},
				},
			}), errUpdate.Err
		}
	}

	if err = pgxTx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
			Response: &pbLegalEntities.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.UpdateAddressResponse{
		Response: &pbLegalEntities.UpdateAddressResponse_Address{
			Address: &pbLegalEntities.Address{
				Id:   legalEntityID,
				Name: "",
				Address: &pbAddresses.LabeledAddress{
					Label:           labelAddress.Label,
					Address:         updatedAddress,
					MainAddress:     labelAddress.MainAddress,
					OwnershipStatus: labelAddress.OwnershipStatus,
					Status:          labelAddress.Status,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) UpdateLegalEntityAddressStatus(
	ctx context.Context,
	tx pgx.Tx,
	legalEntityID string,
	updateLabeledAddressReq *pbAddresses.UpdateLabeledAddressRequest,
) (labelAddress *pbAddresses.LabeledAddress, err error) {
	qbLEAddress, _ := s.legalEntityAddressRepo.QbUpdateLegalEntitiesAddresses(legalEntityID, updateLabeledAddressReq)
	sqlUpsertUserAddress, sqlUpsertUserAddressArgs, _ := qbLEAddress.GenerateSQL()

	rows, err := tx.Query(ctx, sqlUpsertUserAddress, sqlUpsertUserAddressArgs...)
	log.Info().Msgf("query update legal entity address %s with args: %s", sqlUpsertUserAddress, sqlUpsertUserAddressArgs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		labelAddress, err = s.legalEntityAddressRepo.ScanRow(rows)
		if err != nil {
			return nil, err
		}
	}
	return labelAddress, nil
}

func (s *ServiceServer) RemoveAddresses(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.RemoveAddressesRequest],
) (*connect.Response[pbLegalEntities.RemoveAddressesResponse], error) {
	var legalEntityID string
	if errCode, err := s.ValidateRemoveAddresses(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
			Response: &pbLegalEntities.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	pgxTx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
			Response: &pbLegalEntities.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if _err := pgxTx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()

	legalEntityID, err = s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
			Response: &pbLegalEntities.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	qbLEAddresses, err := s.legalEntityAddressRepo.QbRemoveLegalEntitiesAddresses(legalEntityID, req.Msg.GetLabels())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
			Response: &pbLegalEntities.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	sqlUpsertUserAddress, sqlUpsertUserAddressArgs, _ := qbLEAddresses.GenerateSQL()
	log.Info().Msgf("query update legal entity address %s with args: %s", sqlUpsertUserAddress, sqlUpsertUserAddressArgs)
	rows, err := pgxTx.Query(ctx, sqlUpsertUserAddress, sqlUpsertUserAddressArgs...)
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
			Response: &pbLegalEntities.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	defer rows.Close()

	result := []*pbAddresses.LabeledAddress{}
	for rows.Next() {
		labelAddress, err := s.legalEntityAddressRepo.ScanRow(rows)
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
				Response: &pbLegalEntities.RemoveAddressesResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
		result = append(result, labelAddress)
	}

	if err = pgxTx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
			Response: &pbLegalEntities.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.RemoveAddressesResponse{
		Response: &pbLegalEntities.RemoveAddressesResponse_Addresses{
			Addresses: &pbLegalEntities.AddressList{
				Id:   legalEntityID,
				Name: "",
				Addresses: &pbAddresses.LabeledAddressList{
					List: result,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) SetContacts(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.SetContactsRequest],
) (*connect.Response[pbLegalEntities.SetContactsResponse], error) {
	if errCode, err := s.ValidateSetContacts(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
			Response: &pbLegalEntities.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	pgxTx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
			Response: &pbLegalEntities.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if _err := pgxTx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()

	legalEntityID, err := s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
			Response: &pbLegalEntities.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	// get query builder for upsert contacts
	// separate update query and insert query with Contact
	updateContacts, insertContacts := s.SeparateContactsQuery(req.Msg)
	// insert contact first and get id
	res, err := s.InsertMultipleContacts(ctx, pgxTx, req.Msg.GetContacts(), insertContacts)
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
			Response: &pbLegalEntities.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	// query update
	for _, Contact := range updateContacts {
		res[Contact.GetId()] = Contact
	}
	var upsertLabelContacts []*pbContacts.SetLabeledContact
	for k, v := range res {
		labelContact := &pbContacts.SetLabeledContact{
			Label: v.Label,
			Select: &pbContacts.SetLabeledContact_Id{
				Id: k,
			},
			MainContact: v.MainContact,
			Status:      v.Status,
		}
		upsertLabelContacts = append(upsertLabelContacts, labelContact)
	}

	qbUpsertContacts, errUpsert := s.Repo.QbUpsertContacts(legalEntityID, upsertLabelContacts, "upsert")
	if errUpsert != nil {
		return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
			Response: &pbLegalEntities.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    errUpsert.Error(),
				},
			},
		}), nil
	}
	sqlUpsertContacts, sqlUpsertContactsArgs, _ := qbUpsertContacts.GenerateSQL()
	log.Info().Msgf("query upsert contacts %s with args: %s", sqlUpsertContacts, sqlUpsertContactsArgs)
	rowsUpsertContacts, errQueryUpsert := pgxTx.Query(ctx, sqlUpsertContacts, sqlUpsertContactsArgs...)
	if errQueryUpsert != nil {
		return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
			Response: &pbLegalEntities.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    errQueryUpsert.Error(),
				},
			},
		}), nil
	}
	defer rowsUpsertContacts.Close()
	var result []*pbContacts.LabeledContact
	for rowsUpsertContacts.Next() {
		labelContact, err := s.Repo.ScanRow(rowsUpsertContacts)
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
				Response: &pbLegalEntities.SetContactsResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
		result = append(result, labelContact)
	}

	if err = pgxTx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
			Response: &pbLegalEntities.SetContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.SetContactsResponse{
		Response: &pbLegalEntities.SetContactsResponse_Contacts{
			Contacts: &pbLegalEntities.ContactList{
				Id: legalEntityID,
				Contacts: &pbContacts.LabeledContactList{
					List: result,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) InsertMultipleContacts(
	ctx context.Context,
	tx pgx.Tx,
	contactList *pbContacts.SetLabeledContactList,
	insertContacts []*pbContacts.SetLabeledContact,
) (map[string]*pbContacts.SetLabeledContact, error) {
	qbInsertContacts, err := s.contactRepo.QbBulkInsert(contactList)
	if err != nil {
		return nil, err
	}
	sqlInsertContacts, sqlInsertContactsArgs, _ := qbInsertContacts.GenerateSQL()
	log.Info().Msgf("query insert contacts %s with args %s", sqlInsertContacts, sqlInsertContactsArgs)

	rows, err := tx.Query(ctx, sqlInsertContacts, sqlInsertContactsArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	countInsert := 0
	res := map[string]*pbContacts.SetLabeledContact{} // {<Contact_id>: {SetLabelContact}}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		res[id] = insertContacts[countInsert]
		countInsert++
	}
	return res, nil
}

func (s *ServiceServer) AddContacts(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.AddContactsRequest],
) (*connect.Response[pbLegalEntities.AddContactsResponse], error) {
	if errCode, err := s.ValidateAddContacts(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
			Response: &pbLegalEntities.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
			Response: &pbLegalEntities.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	// get user id from req.user
	legalEntityID, err := s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
			Response: &pbLegalEntities.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	// insert contact first and get id
	qbInsertContacts, errInsert := s.contactRepo.QbBulkInsert(req.Msg.GetContacts())
	if errInsert != nil {
		return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
			Response: &pbLegalEntities.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    errInsert.Error(),
				},
			},
		}), nil
	}
	sqlInsertContacts, sqlInsertContactsArgs, _ := qbInsertContacts.GenerateSQL()
	log.Info().Msgf("query insert contacts %s with args %s", sqlInsertContacts, sqlInsertContactsArgs)

	rows, errQuery := tx.Query(ctx, sqlInsertContacts, sqlInsertContactsArgs...)
	if errQuery != nil {
		return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
			Response: &pbLegalEntities.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    errQuery.Error(),
				},
			},
		}), nil
	}
	defer rows.Close()
	countInsert := 0
	insertContacts := req.Msg.GetContacts().GetList()
	res := map[string]*pbContacts.SetLabeledContact{} // {<Contact_id>: {SetLabelContact}}
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
				Response: &pbLegalEntities.AddContactsResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
		res[id] = insertContacts[countInsert]
		countInsert++
	}

	// query update
	var insertLabelContacts []*pbContacts.SetLabeledContact
	for k, v := range res {
		labelContact := &pbContacts.SetLabeledContact{
			Label: v.Label,
			Select: &pbContacts.SetLabeledContact_Id{
				Id: k,
			},
			Status:      v.Status,
			MainContact: v.MainContact,
		}
		insertLabelContacts = append(insertLabelContacts, labelContact)
	}

	result, err := s.UpsertContact(ctx, tx, legalEntityID, insertLabelContacts)
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
			Response: &pbLegalEntities.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	if err = tx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
			Response: &pbLegalEntities.AddContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.AddContactsResponse{
		Response: &pbLegalEntities.AddContactsResponse_Contacts{
			Contacts: &pbLegalEntities.ContactList{
				Id: legalEntityID,
				Contacts: &pbContacts.LabeledContactList{
					List: result,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) UpsertContact(
	ctx context.Context,
	tx pgx.Tx,
	legalEntityID string,
	insertLabelContacts []*pbContacts.SetLabeledContact,
) ([]*pbContacts.LabeledContact, error) {
	qbUpsertContacts, errUpsert := s.Repo.QbUpsertContacts(legalEntityID, insertLabelContacts, "insert")
	if errUpsert != nil {
		return nil, errUpsert
	}
	sqlUpsertContacts, sqlUpsertContactsArgs, _ := qbUpsertContacts.GenerateSQL()
	rowsUpsertContacts, errQueryUpsert := tx.Query(ctx, sqlUpsertContacts, sqlUpsertContactsArgs...)
	log.Info().Msgf("query upsert contacts %s with args: %s", sqlUpsertContacts, sqlUpsertContactsArgs)
	if errQueryUpsert != nil {
		return nil, errQueryUpsert
	}
	defer rowsUpsertContacts.Close()
	var result []*pbContacts.LabeledContact
	for rowsUpsertContacts.Next() {
		labelContact, err := s.Repo.ScanRow(rowsUpsertContacts)
		if err != nil {
			return nil, err
		}
		result = append(result, labelContact)
	}
	return result, nil
}

func (s *ServiceServer) UpdateContact(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.UpdateContactRequest],
) (*connect.Response[pbLegalEntities.UpdateContactResponse], error) {
	if errCode, err := s.ValidateUpdateContact(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
			Response: &pbLegalEntities.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	pgxTx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
			Response: &pbLegalEntities.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if _err := pgxTx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()

	legalEntityID, err := s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
			Response: &pbLegalEntities.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	var (
		updateContact *pbContacts.Contact
		labelContact  *pbContacts.LabeledContact
	)

	if req.Msg.GetContact().Status != nil {
		labelContact, err = s.UpdateLegalEntityContactStatus(ctx, pgxTx, legalEntityID, req.Msg.GetContact())
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
				Response: &pbLegalEntities.UpdateContactResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), nil
		}
	} else {
		qbLabelContact := s.Repo.QbGetOneContact(legalEntityID, req.Msg.GetContact().GetByLabel())
		sqlstr, sqlArgs, sel := qbLabelContact.GenerateSQL()
		log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

		rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
				Response: &pbLegalEntities.UpdateContactResponse_Error{
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
			return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
				Response: &pbLegalEntities.UpdateContactResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			}), _err
		}
		labelContact, err = s.Repo.ScanRow(rows)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
				Response: &pbLegalEntities.UpdateContactResponse_Error{
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
			return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
				Response: &pbLegalEntities.UpdateContactResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			}), _err
		}
	}

	if req.Msg.Contact != nil {
		updateContact, err = s.UpdateContactData(ctx, labelContact.Contact.Id, req.Msg.GetContact().GetContact())
		if err != nil {
			return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
				Response: &pbLegalEntities.UpdateContactResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), err
		}
	}

	if err = pgxTx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
			Response: &pbLegalEntities.UpdateContactResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.UpdateContactResponse{
		Response: &pbLegalEntities.UpdateContactResponse_Contact{
			Contact: &pbLegalEntities.Contact{
				Id: legalEntityID,
				Contact: &pbContacts.LabeledContact{
					Label:       labelContact.Label,
					Contact:     updateContact,
					MainContact: labelContact.MainContact,
					Status:      labelContact.Status,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) UpdateLegalEntityContactStatus(
	ctx context.Context,
	tx pgx.Tx,
	legalEntityID string,
	req *pbContacts.UpdateLabeledContactRequest,
) (labelContact *pbContacts.LabeledContact, err error) {
	qbLEContact, _ := s.Repo.QbUpdateLegalEntitiesContact(legalEntityID, req)

	sqlUpsertContact, sqlUpsertContactArgs, _ := qbLEContact.GenerateSQL()
	log.Info().Msgf("query update legal entity contact %s with args: %s", sqlUpsertContact, sqlUpsertContactArgs)
	rows, err := tx.Query(ctx, sqlUpsertContact, sqlUpsertContactArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		labelContact, err = s.Repo.ScanRow(rows)
		if err != nil {
			return nil, err
		}
	}
	return labelContact, nil
}

func (s *ServiceServer) UpdateContactData(
	ctx context.Context,
	contactID string,
	req *pbContacts.UpdateLabeledContact,
) (*pbContacts.Contact, error) {
	updateContactReq := &pbContacts.UpdateRequest{
		Id:     contactID,
		Type:   req.Type,
		Value:  req.Value,
		Status: req.Status,
	}
	qbContact, err := s.contactRepo.QbUpdate(updateContactReq)
	if err != nil {
		return nil, err
	}

	sqlStr, args, sel := qbContact.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	updateContact, err := common.ExecuteTxWrite[pbContacts.Contact](
		ctx,
		s.db,
		sqlStr,
		args,
		s.contactRepo.ScanRow,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return nil, errUpdate.Err
	}
	return updateContact, nil
}

func (s *ServiceServer) RemoveContacts(
	ctx context.Context,
	req *connect.Request[pbLegalEntities.RemoveContactsRequest],
) (*connect.Response[pbLegalEntities.RemoveContactsResponse], error) {
	var legalEntityID string
	if errCode, err := s.ValidateRemoveContacts(req.Msg); err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
			Response: &pbLegalEntities.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	pgxTx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
			Response: &pbLegalEntities.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if _err := pgxTx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()

	legalEntityID, err = s.GetLegalEntityID(ctx, req.Msg.GetLegalEntity())
	if err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
			Response: &pbLegalEntities.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	qbLEContacts, errRemove := s.Repo.QbRemoveLegalEntitiesContacts(legalEntityID, req.Msg.GetLabels())
	if errRemove != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
			Response: &pbLegalEntities.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    errRemove.Error(),
				},
			},
		}), nil
	}
	sqlUpsertUserContact, sqlUpsertUserContactArgs, _ := qbLEContacts.GenerateSQL()
	log.Info().Msgf("query update legal entity contact %s with args: %s", sqlUpsertUserContact, sqlUpsertUserContactArgs)
	rows, errQuery := pgxTx.Query(ctx, sqlUpsertUserContact, sqlUpsertUserContactArgs...)
	if errQuery != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
			Response: &pbLegalEntities.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    errQuery.Error(),
				},
			},
		}), nil
	}
	defer rows.Close()

	var result []*pbContacts.LabeledContact
	for rows.Next() {
		labelContact, errScan := s.Repo.ScanRow(rows)
		if errScan != nil {
			return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
				Response: &pbLegalEntities.RemoveContactsResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
						Package: _package,
						Text:    errScan.Error(),
					},
				},
			}), nil
		}
		result = append(result, labelContact)
	}

	if err = pgxTx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
			Response: &pbLegalEntities.RemoveContactsResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLegalEntities.RemoveContactsResponse{
		Response: &pbLegalEntities.RemoveContactsResponse_Contacts{
			Contacts: &pbLegalEntities.ContactList{
				Id: legalEntityID,
				Contacts: &pbContacts.LabeledContactList{
					List: result,
				},
			},
		},
	}), nil
}
