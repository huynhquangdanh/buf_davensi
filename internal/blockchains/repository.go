package blockchains

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbBlockchains "davensi.com/core/gen/blockchains"
	pbBlockchainsConnect "davensi.com/core/gen/blockchains/blockchainsconnect"
	pbCommon "davensi.com/core/gen/common"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.blockchains"
	_fields    = "id, name, type, icon, evm, status"
)

type BlockchainRepository struct {
	pbBlockchainsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewBlockchainRepository(db *pgxpool.Pool) *BlockchainRepository {
	return &BlockchainRepository{
		db: db,
	}
}

func (s *BlockchainRepository) QbInsert(msg *pbBlockchains.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleFSProviderValue := []any{}

	// Append required value: type, name, provider
	qb.SetInsertField("name")
	singleFSProviderValue = append(singleFSProviderValue, msg.GetName())

	// Append optional fields values
	if msg.Type != nil {
		qb.SetInsertField("type")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetType())
	}
	if msg.Icon != nil {
		qb.SetInsertField("icon")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetIcon())
	}
	if msg.Evm != nil {
		qb.SetInsertField("evm")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetEvm())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleFSProviderValue)

	return qb, err
}

func (s *BlockchainRepository) QbUpdate(msg *pbBlockchains.UpdateRequest) (qb *util.QueryBuilder, err error) {
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
	if msg.Evm != nil {
		qb.SetUpdate("evm", msg.GetEvm())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetSelect().Select.(type) {
	case *pbBlockchains.Select_ById:
		qb.Where("id = ?", msg.GetSelect().GetById())
	case *pbBlockchains.Select_ByName:
		qb.Where("name = ?", msg.GetSelect().GetByName())
	}

	return qb, nil
}

func (s *BlockchainRepository) QbGetOne(msg *pbBlockchains.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	switch msg.GetSelect().Select.(type) {
	case *pbBlockchains.Select_ById:
		qb.Where("id = ?", msg.GetSelect().GetById())
	case *pbBlockchains.Select_ByName:
		qb.Where("name = ?", msg.GetSelect().GetByName())
	}

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *BlockchainRepository) QbGetList(msg *pbBlockchains.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.Type != nil {
		blockchainTypes := msg.GetType().GetList()

		if len(blockchainTypes) > 0 {
			args := []any{}
			for _, v := range blockchainTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(blockchainTypes)), ""), ", "),
				),
				args...,
			)
		}
	}
	if msg.Name != nil {
		qb.Where("name LIKE '%' || ? || '%'", msg.GetName())
	}
	if msg.Icon != nil {
		qb.Where("icon LIKE '%' || ? || '%'", msg.GetIcon())
	}
	if msg.Evm != nil {
		qb.Where("evm = ?", msg.GetEvm())
	}
	if msg.Status != nil {
		blockchainStatuses := msg.GetStatus().GetList()

		if len(blockchainStatuses) > 0 {
			args := []any{}
			for _, v := range blockchainStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(blockchainStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *BlockchainRepository) ScanRow(row pgx.Row) (*pbBlockchains.Blockchain, error) {
	var (
		id             string
		name           string
		blockchainType pbBlockchains.Type
		icon           sql.NullString
		evm            bool
		status         pbCommon.Status
	)

	err := row.Scan(
		&id,
		&name,
		&blockchainType,
		&icon,
		&evm,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbBlockchains.Blockchain{
		Id:     id,
		Name:   name,
		Type:   blockchainType,
		Icon:   util.GetSQLNullString(icon),
		Evm:    evm,
		Status: status,
	}, nil
}
