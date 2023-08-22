package orgs

import (
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbOrgs "davensi.com/core/gen/orgs"
	pbOrgsConnect "davensi.com/core/gen/orgs/orgsconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName  = "core.orgs"
	_orgsFields = "id, name, type, status"
)

type OrgRepository struct {
	pbOrgsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewOrgRepository(db *pgxpool.Pool) *OrgRepository {
	return &OrgRepository{
		db: db,
	}
}

func (s *OrgRepository) QbInsert(msg *pbOrgs.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleOrgValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("name")
	singleOrgValue = append(singleOrgValue, msg.GetName())

	// Append optional fields values
	if msg.Type != nil {
		qb.SetInsertField("type")
		singleOrgValue = append(singleOrgValue, msg.GetType())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleOrgValue = append(singleOrgValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleOrgValue)

	return qb, err
}

func (s *OrgRepository) QbUpdate(msg *pbOrgs.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
	}

	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetSelect().(type) {
	case *pbOrgs.UpdateRequest_ById:
		qb.Where("id = ?", msg.GetById())
	case *pbOrgs.UpdateRequest_ByName:
		qb.Where("name = ?", msg.GetByName())
	}

	return qb, nil
}

func (s *OrgRepository) QbGetOne(msg *pbOrgs.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_orgsFields)

	switch msg.GetSelect().(type) {
	case *pbOrgs.GetRequest_ById:
		qb.Where("id = ?", msg.GetById())
	case *pbOrgs.GetRequest_ByName:
		qb.Where("name = ?", msg.GetByName())
	}

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *OrgRepository) QbGetList(req *pbOrgs.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_orgsFields, "orgs"))

	if req.Name != nil {
		qb.Where("orgs.name LIKE '%' || ? || '%'", req.GetName())
	}

	if req.Type != nil {
		orgTypes := req.GetType().GetList()

		if len(orgTypes) > 0 {
			args := []any{}
			for _, v := range orgTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"orgs.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(orgTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if req.Status != nil {
		orgStatuses := req.GetStatus().GetList()

		if len(orgStatuses) > 0 {
			args := []any{}
			for _, v := range orgStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"orgs.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(orgStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *OrgRepository) QbDelete(req *pbOrgs.DeleteRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Delete, _tableName)
	qb.Select(_orgsFields)

	switch req.GetSelect().(type) {
	case *pbOrgs.DeleteRequest_ById:
		qb.Where("id = ?", req.GetById())
	case *pbOrgs.DeleteRequest_ByName:
		qb.Where("name = ?", req.GetByName())
	}

	return qb
}

func ScanRow(row pgx.Row) (*pbOrgs.Org, error) {
	var (
		id      string
		name    string
		orgType uint32
		status  pbCommon.Status
	)

	err := row.Scan(
		&id,
		&name,
		&orgType,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbOrgs.Org{
		Id:     id,
		Name:   name,
		Type:   pbOrgs.Type(orgType),
		Status: status,
	}, nil
}
