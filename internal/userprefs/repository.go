package userprefs

import (
	"context"
	"errors"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbUserPrefs "davensi.com/core/gen/userprefs"
	pbUserPrefsConnect "davensi.com/core/gen/userprefs/userprefsconnect"
	pbUsers "davensi.com/core/gen/users"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName       = "core.userprefs"
	_userPrefsFields = "user_id, key, value, status"
	_usersFields     = "id, login, type, screen_name, avatar, status"
	_joinfields      = "users.id, users.type, users.screen_name, users.avatar, users.status as ust, key, value, status as pst"
)

type UserPrefsRepository struct {
	pbUserPrefsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewUserPrefRepository(db *pgxpool.Pool) *UserPrefsRepository {
	return &UserPrefsRepository{
		db: db,
	}
}

func (s *UserPrefsRepository) QbSetInsert(req *pbUserPrefs.SetRequest, userID string) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleUserPrefValue := []any{}

	// Append required value: user_id, key
	qb.SetInsertField("user_id")
	// switch req.User.Select.(type) {	// MAY BE IMPLMENTED LATER
	// MAY BE POSSIBLE TO GET user_id directly from repo?
	// case *pbUsers.Select_ById:
	// 	singleUserPrefValue = append(singleUserPrefValue,
	// 		fmt.Sprintf("(SELECT id FROM core.users WHERE id = '%s')", req.GetUser().GetById()))
	// case *pbUsers.Select_ByLogin:
	// 	singleUserPrefValue = append(singleUserPrefValue,
	// 		fmt.Sprintf("(SELECT id FROM core.users WHERE login = '%s')", req.GetUser().GetByLogin()))
	// }
	singleUserPrefValue = append(singleUserPrefValue, userID)
	qb.SetInsertField("key")
	singleUserPrefValue = append(singleUserPrefValue, req.GetKey())

	// Append optional fields values
	qb.SetInsertField("value")
	if req.Value != nil {
		singleUserPrefValue = append(singleUserPrefValue, req.GetValue())
	} else {
		singleUserPrefValue = append(singleUserPrefValue,
			fmt.Sprintf("(SELECT value from core.userprefs_default WHERE key = '%s')", req.GetKey()))
	}
	qb.SetInsertField("status")
	singleUserPrefValue = append(singleUserPrefValue, pbCommon.Status_STATUS_ACTIVE)

	_, err := qb.SetInsertValues(singleUserPrefValue)

	return qb, err
}

func (s *UserPrefsRepository) ExecuteQueryAndReturn(ctx context.Context, sql string, args ...string) (pgx.Rows, error) {
	rows, err := s.db.Query(ctx, sql, args)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *UserPrefsRepository) QbSetUpdate(req *pbUserPrefs.SetRequest, userID string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if req.Value != nil {
		qb.SetUpdate("value", req.GetValue())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	// switch req.GetUser().Select.(type) {	// MAY IMPLEMENT LATER
	// MAY BE POSSIBLE TO GET user_id directly from repo?
	// case *pbUsers.Select_ById:
	// 	qb.Where("user_id = ?",
	// 		fmt.Sprintf("(SELECT id FROM core.users WHERE id = '%s')", req.GetUser().GetById()))
	// case *pbUsers.Select_ByLogin:
	// 	qb.Where("user_id = ?",
	// 		fmt.Sprintf("(SELECT id FROM core.users WHERE login = '%s')", req.GetUser().GetByLogin()))
	// }

	qb.Where("user_id = ?", userID)
	qb.Where("key = ?", req.GetKey())

	return qb, nil
}

func (s *UserPrefsRepository) QbGetOne(req *pbUserPrefs.GetRequest, userID string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)

	qb.Select(_joinfields)
	qb.Join(fmt.Sprintf("core.users ON users.id = %s.user_id", _package))
	qb.Where("user_id = ?", userID)
	qb.Where("key = ?", req.GetKey())

	return qb
}

func (s *UserPrefsRepository) QbGetList(req *pbUserPrefs.GetListRequest, userID string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_joinfields)
	qb.Join(fmt.Sprintf("core.users ON users.id = %s.user_id", _package))

	qb.Where(fmt.Sprintf("%s.user_id = ?", _package), userID)

	if req.GetKeyPrefix() != "" {
		qb.Where("userprefs.key LIKE '%' || ? || '%'", req.GetKeyPrefix())
	}

	return qb
}

func (s *UserPrefsRepository) QbRemove(req *pbUserPrefs.RemoveRequest, userID string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetUpdate("status", pbCommon.Status_STATUS_TERMINATED)

	qb.Where("user_id = ?", userID)
	qb.Where("key = ?", req.GetKey())

	return qb, nil
}

func (s *UserPrefsRepository) QbReset(req *pbUserPrefs.ResetRequest, userID string) (qb *util.QueryBuilder) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetUpdate("status", pbCommon.Status_STATUS_TERMINATED)

	qb.Where("user_id = ?", userID)

	if req.GetKeyPrefix() != "" {
		qb.Where("key LIKE '%' || ? || '%'", req.GetKeyPrefix())
	}

	return qb
}

func (s *UserPrefsRepository) ScanRowWithRelationship(row pgx.Row) (*pbUserPrefs.UserPref, error) {
	// users.id, users.type, users.screen_name, users.avatar, users.status as ust, key, value, status as pst
	var (
		userID     string
		userType   uint32
		screenName string
		avatar     string
		userStatus uint32
		key        string
		value      string
		status     uint32
	)

	err := row.Scan(
		&userID,
		&userType,
		&screenName,
		&avatar,
		&userStatus,
		&key,
		&value,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbUserPrefs.UserPref{
		User: &pbUsers.User{
			Id:         userID,
			Type:       pbUsers.Type(userType),
			ScreenName: &screenName,
			Avatar:     &avatar,
			Status:     pbCommon.Status(userStatus),
		},
		Key:    key,
		Value:  value,
		Status: pbCommon.Status(status),
	}, nil
}

func ScanRow(row pgx.Row) (*pbUserPrefs.UserPref, error) {
	var (
		userID string
		key    string
		value  string
		status uint32
	)

	err := row.Scan(
		&userID,
		&key,
		&value,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbUserPrefs.UserPref{
		User: &pbUsers.User{
			Id: userID,
		},
		Key:    key,
		Value:  value,
		Status: pbCommon.Status(status),
	}, nil
}
