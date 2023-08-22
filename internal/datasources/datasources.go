package datasources

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbDataSourcesConnect "davensi.com/core/gen/datasources/datasourcesconnect"
	pbFSProviders "davensi.com/core/gen/fsproviders"

	fsproviders "davensi.com/core/internal/fsproviders"
)

const (
	_package          = "datasources"
	_entityName       = "Data Source"
	_entityNamePlural = "Data Sources"
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	Repo DataSourceRepository
	pbDataSourcesConnect.UnimplementedServiceHandler
	db         *pgxpool.Pool
	providerSS *fsproviders.ServiceServer
}

// For singleton UoMs export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:       *NewDataSourceRepository(db),
		providerSS: fsproviders.GetSingletonServiceServer(db),
		db:         db,
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbDataSources.CreateRequest],
) (*connect.Response[pbDataSources.CreateResponse], error) {
	if errCreation := s.validateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbDataSources.CreateResponse{
			Response: &pbDataSources.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"create",
			_package,
			err.Error(),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbDataSources.CreateResponse{
			Response: &pbDataSources.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")

	newDataSource, err := common.ExecuteTxWrite(
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanMainEntity,
	)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			fmt.Sprintf(
				"type/name = '%s/%s'",
				req.Msg.GetType().Enum().String(),
				req.Msg.GetName(),
			),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbDataSources.CreateResponse{
			Response: &pbDataSources.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s with type/name = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetType().Enum().String()+"/"+req.Msg.GetName(), newDataSource.Id)
	return connect.NewResponse(&pbDataSources.CreateResponse{
		Response: &pbDataSources.CreateResponse_Datasource{
			Datasource: newDataSource,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbDataSources.UpdateRequest],
) (*connect.Response[pbDataSources.UpdateResponse], error) {
	if errQueryUpdate := s.validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbDataSources.UpdateResponse{
			Response: &pbDataSources.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getDataSourceResponse, err := s.Get(context.Background(), &connect.Request[pbDataSources.GetRequest]{
		Msg: &pbDataSources.GetRequest{
			Select: req.Msg.Select,
		},
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbDataSources.UpdateResponse{
			Response: &pbDataSources.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getDataSourceResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getDataSourceResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	dataSourceBeforeUpdate := getDataSourceResponse.Msg.GetDatasource()

	if errUpdateValue := s.validateUpdateValue(dataSourceBeforeUpdate, req.Msg); errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbDataSources.UpdateResponse{
			Response: &pbDataSources.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdateValue.Code,
					Package: _package,
					Text:    errUpdateValue.Err.Error(),
				},
			},
		}), errUpdateValue.Err
	}

	qb, err := s.Repo.QbUpdate(req.Msg)
	if err != nil {
		errGenSQL := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			err.Error(),
		)
		log.Error().Err(errGenSQL.Err)
		return connect.NewResponse(&pbDataSources.UpdateResponse{
			Response: &pbDataSources.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGenSQL.Code,
					Package: _package,
					Text:    errGenSQL.Err.Error(),
				},
			},
		}), errGenSQL.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedDataSource, err := common.ExecuteTxWrite[pbDataSources.DataSource](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanMainEntity,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbDataSources.UpdateResponse{
			Response: &pbDataSources.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbDataSources.UpdateResponse{
		Response: &pbDataSources.UpdateResponse_Datasource{
			Datasource: updatedDataSource,
		},
	}), nil
}

func (s *ServiceServer) GetMainEntity(
	ctx context.Context,
	req *connect.Request[pbDataSources.GetRequest],
) (*connect.Response[pbDataSources.GetResponse], error) {
	if errQueryGet := ValidateSelect(req.Msg.Select, "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
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
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, errExcute := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errExcute != nil {
		queryErr := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"fetching",
			_entityName,
			fmt.Sprintf("%s %s", errExcute.Error(), sel),
		)
		log.Error().Err(queryErr.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    queryErr.Code,
					Package: _package,
					Text:    queryErr.Err.Error(),
				},
			},
		}), queryErr.Err
	}

	defer rows.Close()

	if !rows.Next() {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
			"fetching",
			_entityName,
			fmt.Sprintf("not found %s", sel),
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	dataSource, errExcute := s.Repo.ScanMainEntity(rows)
	if errExcute != nil {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
			"fetching",
			_package,
			fmt.Sprintf("%s, %s", sel, errExcute.Error()),
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	if rows.Next() {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND,
			"fetching",
			_entityNamePlural,
			sel,
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbDataSources.GetResponse{
		Response: &pbDataSources.GetResponse_Datasource{
			Datasource: dataSource,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbDataSources.GetRequest],
) (*connect.Response[pbDataSources.GetResponse], error) {
	if errQueryGet := ValidateSelect(req.Msg.Select, "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qbProviders := s.providerSS.Repo.QbGetList(&pbFSProviders.GetListRequest{})
	providerFB, providerArgs := qbProviders.Filters.GenerateSQL()

	qb := s.Repo.QbGetOne(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON fsproviders.id = datasources.fsprovider_id", qbProviders.TableName)).
		Select(strings.Join(qbProviders.SelectFields, ", ")).
		Where(providerFB, providerArgs...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, errExcute := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errExcute != nil {
		queryErr := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"fetching",
			_entityName,
			fmt.Sprintf("%s %s", errExcute.Error(), sel),
		)
		log.Error().Err(queryErr.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    queryErr.Code,
					Package: _package,
					Text:    queryErr.Err.Error(),
				},
			},
		}), queryErr.Err
	}

	defer rows.Close()

	if !rows.Next() {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
			"fetching",
			_entityName,
			fmt.Sprintf("not found %s", sel),
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	dataSource, errExcute := s.Repo.ScanWithRelationship(rows)
	if errExcute != nil {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
			"fetching",
			_package,
			fmt.Sprintf("%s, %s", sel, errExcute.Error()),
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	if rows.Next() {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND,
			"fetching",
			_entityNamePlural,
			sel,
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbDataSources.GetResponse{
			Response: &pbDataSources.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbDataSources.GetResponse{
		Response: &pbDataSources.GetResponse_Datasource{
			Datasource: dataSource,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbDataSources.GetListRequest],
	res *connect.ServerStream[pbDataSources.GetListResponse],
) error {
	if req.Msg.Provider == nil {
		req.Msg.Provider = &pbFSProviders.GetListRequest{}
	}
	qbProviders := s.providerSS.Repo.QbGetList(req.Msg.Provider)
	providerFB, providerArgs := qbProviders.Filters.GenerateSQL()

	qb := s.Repo.QbGetList(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON fsproviders.id = datasources.fsprovider_id", qbProviders.TableName)).
		Select(strings.Join(qbProviders.SelectFields, ", ")).
		Where(providerFB, providerArgs...)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbDataSources.GetListResponse{
					Response: &pbDataSources.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		dataSource, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbDataSources.GetListResponse{
						Response: &pbDataSources.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbDataSources.GetListResponse{
			Response: &pbDataSources.GetListResponse_Datasource{
				Datasource: dataSource,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbDataSources.DeleteRequest],
) (*connect.Response[pbDataSources.DeleteResponse], error) {
	deletedDataSource, err := s.Update(ctx, connect.NewRequest(&pbDataSources.UpdateRequest{
		Select: req.Msg.Select,
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbDataSources.DeleteResponse{
			Response: &pbDataSources.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedDataSource.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedDataSource.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbDataSources.DeleteResponse{
		Response: &pbDataSources.DeleteResponse_Datasource{
			Datasource: deletedDataSource.Msg.GetDatasource(),
		},
	}), nil
}
