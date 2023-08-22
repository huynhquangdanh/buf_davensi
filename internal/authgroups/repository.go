package authgroups

import (
	"errors"
	"fmt"
	"strings"

	pbAuthGroups "davensi.com/core/gen/authgroups"
	pbAuthGroupsConnect "davensi.com/core/gen/authgroups/authgroupsconnect"
	pbCommon "davensi.com/core/gen/common"
	"davensi.com/core/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.authgroups"
	_fields    = "id, name, status"
)

type AuthGroupRepository struct {
	pbAuthGroupsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewAuthGroupRepository(db *pgxpool.Pool) *AuthGroupRepository {
	return &AuthGroupRepository{
		db: db,
	}
}

func (s *AuthGroupRepository) QbInsert(msg *pbAuthGroups.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleAuthGroupValue := []any{}

	// Append required value: name
	qb.SetInsertField("name")
	singleAuthGroupValue = append(singleAuthGroupValue, msg.GetName())

	// Append optional fields values
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleAuthGroupValue = append(singleAuthGroupValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleAuthGroupValue)

	return qb, err
}

func (s *AuthGroupRepository) QbUpdate(msg *pbAuthGroups.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetSelect().Select.(type) {
	case *pbAuthGroups.Select_ById:
		qb.Where("id = ?", msg.GetSelect().GetById())
	case *pbAuthGroups.Select_ByName:
		qb.Where("name = ?", msg.GetSelect().GetByName())
	}

	return qb, nil
}

func (s *AuthGroupRepository) QbGetOne(msg *pbAuthGroups.GetRequest, status *pbCommon.Status) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	switch msg.GetSelect().Select.(type) {
	case *pbAuthGroups.Select_ById:
		qb.Where("id = ?", msg.GetSelect().GetById())
	case *pbAuthGroups.Select_ByName:
		qb.Where("name = ?", msg.GetSelect().GetByName())
	}

	if status != nil {
		qb.Where("status = ?", status)
	}

	return qb
}

func (s *AuthGroupRepository) QbGetList(msg *pbAuthGroups.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.Name != nil {
		qb.Where("name LIKE '%' || ? || '%'", msg.GetName())
	}
	if msg.Status != nil {
		authGroupStatuses := msg.GetStatus().GetList()

		if len(authGroupStatuses) > 0 {
			args := []any{}
			for _, v := range authGroupStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(authGroupStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *AuthGroupRepository) ScanRow(row pgx.Row) (*pbAuthGroups.AuthGroup, error) {
	var (
		id     string
		name   string
		status pbCommon.Status
	)

	err := row.Scan(
		&id,
		&name,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbAuthGroups.AuthGroup{
		Id:     id,
		Name:   name,
		Status: status,
	}, nil
}
