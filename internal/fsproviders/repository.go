package fsproviders

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbFSProviders "davensi.com/core/gen/fsproviders"
	pbFSProvidersConnect "davensi.com/core/gen/fsproviders/fsprovidersconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.fsproviders"
	_fields    = "id, type, name, icon, status"
)

func SetQBBySelect(selectProvider *pbFSProviders.Select, qb *util.QueryBuilder) {
	switch selectProvider.GetSelect().(type) {
	case *pbFSProviders.Select_ById:
		qb.Where("fsproviders.id = ?", selectProvider.GetById())
	case *pbFSProviders.Select_ByTypeName:
		qb.Where(
			"fsproviders.type = ? AND fsproviders.name = ?",
			selectProvider.GetByTypeName().GetType(),
			selectProvider.GetByTypeName().GetName(),
		)
	}
}

type FSProviderRepository struct {
	pbFSProvidersConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewFSProviderRepository(db *pgxpool.Pool) *FSProviderRepository {
	return &FSProviderRepository{
		db: db,
	}
}

func (s *FSProviderRepository) QbInsert(msg *pbFSProviders.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleFSProviderValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("type").SetInsertField("name")
	singleFSProviderValue = append(singleFSProviderValue, msg.GetType(), msg.GetName())

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

func (s *FSProviderRepository) QbUpdate(msg *pbFSProviders.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
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

func (s *FSProviderRepository) QbGetOne(msg *pbFSProviders.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, "fsproviders"))

	SetQBBySelect(msg.Select, qb)

	qb.Where("fsproviders.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *FSProviderRepository) QbGetList(msg *pbFSProviders.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, "fsproviders"))

	if msg.Type != nil {
		fsproviderTypes := msg.GetType().GetList()

		if len(fsproviderTypes) > 0 {
			args := []any{}
			for _, v := range fsproviderTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"fsproviders.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(fsproviderTypes)), ""), ", "),
				),
				args...,
			)
		}
	}
	if msg.Name != nil {
		qb.Where("fsproviders.name LIKE '%' || ? || '%'", msg.GetName())
	}
	if msg.Icon != nil {
		qb.Where("fsproviders.icon LIKE '%' || ? || '%'", msg.GetIcon())
	}
	if msg.Status != nil {
		fsproviderStatuses := msg.GetStatus().GetList()

		if len(fsproviderStatuses) > 0 {
			args := []any{}
			for _, v := range fsproviderStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"fsproviders.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(fsproviderStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *FSProviderRepository) ScanRow(row pgx.Row) (*pbFSProviders.FSProvider, error) {
	var (
		fsproviderID     string
		fsproviderType   pbFSProviders.Type
		fsproviderName   string
		fsproviderIcon   sql.NullString
		fsproviderStatus pbCommon.Status
	)

	err := row.Scan(
		&fsproviderID,
		&fsproviderType,
		&fsproviderName,
		&fsproviderIcon,
		&fsproviderStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbFSProviders.FSProvider{
		Id:     fsproviderID,
		Type:   fsproviderType,
		Name:   fsproviderName,
		Icon:   util.GetSQLNullString(fsproviderIcon),
		Status: fsproviderStatus,
	}, nil
}
