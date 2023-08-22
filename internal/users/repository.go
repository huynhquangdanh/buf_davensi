package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbUsers "davensi.com/core/gen/users"
	pbUsersConnect "davensi.com/core/gen/users/usersconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName   = "core.users"
	_usersFields = "id, login, type, screen_name, avatar, status"
)

type UserRepository struct {
	pbUsersConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (s *UserRepository) QbInsert(req *pbUsers.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleUserValue := []any{}

	// Append required value: login
	qb.SetInsertField("login")
	singleUserValue = append(singleUserValue, req.GetLogin())

	// Append optional fields values
	if req.Type != nil {
		qb.SetInsertField("type")
		singleUserValue = append(singleUserValue, req.GetType())
	}
	if req.ScreenName != nil {
		qb.SetInsertField("screen_name")
		singleUserValue = append(singleUserValue, req.GetScreenName())
	}
	if req.Avatar != nil {
		qb.SetInsertField("avatar")
		singleUserValue = append(singleUserValue, req.GetAvatar())
	}
	if req.Status != nil {
		qb.SetInsertField("status")
		singleUserValue = append(singleUserValue, req.GetStatus())
	}

	_, err := qb.SetInsertValues(singleUserValue)

	return qb, err
}

func (s *UserRepository) QbUpdate(req *pbUsers.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if req.Login != nil {
		qb.SetUpdate("login", req.GetLogin())
	}

	if req.Type != nil {
		qb.SetUpdate("type", req.GetType())
	}

	if req.ScreenName != nil {
		qb.SetUpdate("screen_name", req.GetScreenName())
	}

	if req.Avatar != nil {
		qb.SetUpdate("avatar", req.GetAvatar())
	}

	if req.Status != nil {
		qb.SetUpdate("status", req.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch req.GetSelect().(type) {
	case *pbUsers.UpdateRequest_ById:
		qb.Where("id = ?", req.GetById())
	case *pbUsers.UpdateRequest_ByLogin:
		qb.Where(
			"login = ?", req.GetByLogin(),
		)
	}

	return qb, nil
}

func (s *UserRepository) QbGetOne(req *pbUsers.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_usersFields)

	switch req.GetSelect().(type) {
	case *pbUsers.GetRequest_ById:
		qb.Where("id = ?", req.GetById())
	case *pbUsers.GetRequest_ByLogin:
		qb.Where("login = ?", req.GetByLogin())
	}

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *UserRepository) QbGetList(req *pbUsers.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_usersFields, "users"))

	if req.Login != nil {
		qb.Where("users.login LIKE '%' || ? || '%'", req.GetLogin())
	}

	if req.Type != nil {
		userTypes := req.GetType().GetList()

		if len(userTypes) > 0 {
			args := []any{}
			for _, v := range userTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"users.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(userTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if req.ScreenName != nil {
		qb.Where("users.screen_name LIKE '%' || ? || '%'", req.GetScreenName())
	}
	if req.Avatar != nil {
		qb.Where("users.avatar LIKE '%' || ? || '%'", req.GetAvatar())
	}

	if req.Status != nil {
		userStatuses := req.GetStatus().GetList()

		if len(userStatuses) > 0 {
			args := []any{}
			for _, v := range userStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"users.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(userStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *UserRepository) QbDelete(req *pbUsers.DeleteRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Delete, _tableName)
	qb.Select(_usersFields)

	switch req.GetSelect().(type) {
	case *pbUsers.DeleteRequest_ById:
		qb.Where("id = ?", req.GetById())
	case *pbUsers.DeleteRequest_ByLogin:
		qb.Where("login = ?", req.GetByLogin())
	}

	return qb
}

func (s *UserRepository) ScanRow(row pgx.Row) (*pbUsers.User, error) {
	var (
		id                 string
		login              string
		userType           pbUsers.Type
		nullableScreenName sql.NullString
		nullableAvatar     sql.NullString
		status             pbCommon.Status
	)

	err := row.Scan(
		&id,
		&login,
		&userType,
		&nullableScreenName,
		&nullableAvatar,
		&status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing
	var avatar, screenName *string
	if nullableScreenName.Valid {
		screenName = &nullableScreenName.String
	}
	if nullableAvatar.Valid {
		avatar = &nullableAvatar.String
	}

	return &pbUsers.User{
		Id:         id,
		Login:      login,
		Type:       userType,
		ScreenName: screenName,
		Avatar:     avatar,
		Status:     status,
	}, nil
}

func SetQBBySelect(selectUser *pbUsers.Select, qb *util.QueryBuilder) {
	switch selectUser.GetSelect().(type) {
	case *pbUsers.Select_ById:
		qb.Where("users.id = ?", selectUser.GetById())
	case *pbUsers.Select_ByLogin:
		qb.Where("users.login = ?", selectUser.GetByLogin())
	}
}
