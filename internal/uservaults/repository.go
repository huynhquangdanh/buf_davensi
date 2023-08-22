package uservaults

import (
	"errors"

	pbCommon "davensi.com/core/gen/common"
	pbUserVaults "davensi.com/core/gen/uservaults"
	pbUserVaultsConnect "davensi.com/core/gen/uservaults/uservaultsconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName        = "core.uservaults"
	_userVaultsFields = "user_id, key, value_type, data, status"
	_usersFields      = "id, login, type, screen_name, avatar, status"
)

type UserVaultsRepository struct {
	pbUserVaultsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewUserVaultRepository(db *pgxpool.Pool) *UserVaultsRepository {
	return &UserVaultsRepository{
		db: db,
	}
}

func (s *UserVaultsRepository) QbSetInsert(req *pbUserVaults.SetRequest, userID, hexDataString string, valueType pbUserVaults.ValueType,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleUserPrefValue := []any{}

	qb.SetInsertField("user_id")
	singleUserPrefValue = append(singleUserPrefValue, userID)
	qb.SetInsertField("key")
	singleUserPrefValue = append(singleUserPrefValue, req.GetKey())
	qb.SetInsertField("value_type")
	singleUserPrefValue = append(singleUserPrefValue, valueType)
	qb.SetInsertField("data")
	singleUserPrefValue = append(singleUserPrefValue, hexDataString)
	qb.SetInsertField("status")
	singleUserPrefValue = append(singleUserPrefValue, pbCommon.Status_STATUS_ACTIVE)

	_, err := qb.SetInsertValues(singleUserPrefValue)

	return qb, err
}

func (s *UserVaultsRepository) QbSetUpdate(req *pbUserVaults.SetRequest, userID, hexDataString string, valueType pbUserVaults.ValueType,
) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetUpdate("data", hexDataString)
	qb.SetUpdate("value_type", valueType)

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("user_id = ?", userID)
	qb.Where("key = ?", req.GetKey())

	return qb, nil
}

func (s *UserVaultsRepository) QbGetOne(req *pbUserVaults.GetRequest, userID string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)

	qb.Select(_userVaultsFields)
	qb.Where("user_id = ?", userID)
	qb.Where("key = ?", req.GetKey())

	return qb
}

func (s *UserVaultsRepository) QbGetList(req *pbUserVaults.GetListRequest, userID string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_userVaultsFields)

	qb.Where("user_id = ?", userID)

	if req.GetKeyPrefix() != "" {
		qb.Where("key LIKE '%' || ? || '%'", req.GetKeyPrefix())
	}

	return qb
}

func (s *UserVaultsRepository) QbRemove(req *pbUserVaults.RemoveRequest, userID string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetUpdate("status", pbCommon.Status_STATUS_TERMINATED)

	qb.Where("user_id = ?", userID)
	qb.Where("key = ?", req.GetKey())

	return qb, nil
}

func (s *UserVaultsRepository) QbReset(req *pbUserVaults.ResetRequest, userID string) (qb *util.QueryBuilder) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetUpdate("status", pbCommon.Status_STATUS_TERMINATED)

	qb.Where("user_id = ?", userID)

	if req.GetKeyPrefix() != "" {
		qb.Where("key LIKE '%' || ? || '%'", req.GetKeyPrefix())
	}

	return qb
}

func ScanRowValue(row pgx.Row) (dataByte []byte, dataType uint32, err error) {
	var (
		userID    string
		key       string
		valueType uint32
		data      pgtype.Bytea
		status    uint32
	)

	err = row.Scan(
		&userID,
		&key,
		&valueType,
		&data,
		&status,
	)
	if err != nil {
		return nil, 0, err
	}

	return data.Bytes, valueType, nil
}

func ScanRow(row pgx.Row) (dataByte []byte, dataType uint32, key string, err error) {
	var (
		userID string
		// key       string
		valueType uint32
		data      pgtype.Bytea
		status    uint32
	)

	err = row.Scan(
		&userID,
		&key,
		&valueType,
		&data,
		&status,
	)
	if err != nil {
		return nil, 0, "", err
	}

	return data.Bytes, valueType, key, nil
}
