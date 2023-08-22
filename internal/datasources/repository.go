package datasources

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbDataSourcesConnect "davensi.com/core/gen/datasources/datasourcesconnect"
	pbFsproviders "davensi.com/core/gen/fsproviders"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName       = "core.datasources"
	DataSourceFields = "id, type, name, fsprovider_id, icon, status"
)

type DataSourceRepository struct {
	pbDataSourcesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewDataSourceRepository(db *pgxpool.Pool) *DataSourceRepository {
	return &DataSourceRepository{
		db: db,
	}
}

func SetQBBySelect(selectDatasource *pbDataSources.Select, qb *util.QueryBuilder) {
	switch selectDatasource.GetSelect().(type) {
	case *pbDataSources.Select_ById:
		qb.Where("datasources.id = ?", selectDatasource.GetById())
	case *pbDataSources.Select_ByTypeName:
		qb.Where(
			"datasources.type = ? AND datasources.name = ?",
			selectDatasource.GetByTypeName().GetType(),
			selectDatasource.GetByTypeName().GetName(),
		)
	}
}

func (s *DataSourceRepository) QbInsert(msg *pbDataSources.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleFSProviderValue := []any{}

	// Append required value: type, name, provider
	qb.SetInsertField("type").SetInsertField("name").SetInsertField("fsprovider_id")
	singleFSProviderValue = append(singleFSProviderValue, msg.GetType(), msg.GetName(), msg.GetProvider().GetById())

	// Append optional fields values
	if msg.Icon != nil {
		qb.SetInsertField("icon")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetIcon())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleFSProviderValue)

	return qb, err
}

func (s *DataSourceRepository) QbUpdate(msg *pbDataSources.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}
	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
	}
	if msg.Provider != nil {
		qb.SetUpdate("fsprovider_id", msg.GetProvider().GetById())
	}
	if msg.Icon != nil {
		qb.SetUpdate("icon", msg.GetIcon())
	}
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}
	SetQBBySelect(msg.Select, qb)
	return qb, nil
}

func (s *DataSourceRepository) QbGetOne(msg *pbDataSources.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(DataSourceFields, _tableName))

	SetQBBySelect(msg.Select, qb)

	qb.Where("datasources.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *DataSourceRepository) QbGetList(msg *pbDataSources.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(DataSourceFields, "datasources"))

	if msg.Type != nil {
		dataSourceTypes := msg.GetType().GetList()

		if len(dataSourceTypes) > 0 {
			args := []any{}
			for _, v := range dataSourceTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"datasources.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(dataSourceTypes)), ""), ", "),
				),
				args...,
			)
		}
	}
	if msg.Name != nil {
		qb.Where("datasources.name LIKE '%' || ? || '%'", msg.GetName())
	}
	// TODO: add provider condition when implement relationship
	if msg.Icon != nil {
		qb.Where("datasources.icon LIKE '%' || ? || '%'", msg.GetIcon())
	}
	if msg.Status != nil {
		dataSourceStatuses := msg.GetStatus().GetList()

		if len(dataSourceStatuses) > 0 {
			args := []any{}
			for _, v := range dataSourceStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"datasources.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(dataSourceStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *DataSourceRepository) ScanMainEntity(row pgx.Row) (*pbDataSources.DataSource, error) {
	var (
		dsID           string
		dsType         pbDataSources.Type
		dsName         string
		dsFsProviderID sql.NullString
		dsIcon         sql.NullString
		dsStatus       pbCommon.Status
	)

	err := row.Scan(
		&dsID,
		&dsType,
		&dsName,
		&dsFsProviderID,
		&dsIcon,
		&dsStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbDataSources.DataSource{
		Id:       dsID,
		Type:     dsType,
		Name:     dsName,
		Provider: &pbFsproviders.FSProvider{Id: util.GetPointString(util.GetSQLNullString(dsFsProviderID))},
		Icon:     util.GetSQLNullString(dsIcon),
		Status:   dsStatus,
	}, nil
}

func (s *DataSourceRepository) ScanWithRelationship(row pgx.Row) (*pbDataSources.DataSource, error) {
	var (
		dsID           string
		dsType         pbDataSources.Type
		dsName         string
		dsFsProviderID sql.NullString
		dsIcon         sql.NullString
		dsStatus       pbCommon.Status
	)

	var (
		fsproviderID     string
		fsproviderType   pbFsproviders.Type
		fsproviderName   string
		fsproviderIcon   sql.NullString
		fsproviderStatus pbCommon.Status
	)

	err := row.Scan(
		&dsID,
		&dsType,
		&dsName,
		&dsFsProviderID,
		&dsIcon,
		&dsStatus,
		&fsproviderID,
		&fsproviderType,
		&fsproviderName,
		&fsproviderIcon,
		&fsproviderStatus,
	)
	if err != nil {
		return nil, err
	}

	fmt.Println(fsproviderID, fsproviderType, fsproviderName)

	return &pbDataSources.DataSource{
		Id:   dsID,
		Type: dsType,
		Name: dsName,
		Provider: &pbFsproviders.FSProvider{
			Id:     fsproviderID,
			Type:   fsproviderType,
			Name:   fsproviderName,
			Icon:   util.GetSQLNullString(fsproviderIcon),
			Status: fsproviderStatus,
		},
		Icon:   util.GetSQLNullString(dsIcon),
		Status: dsStatus,
	}, nil
}
