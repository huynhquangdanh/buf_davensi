package banks

import (
	"errors"
	"fmt"
	"strings"

	pbBanks "davensi.com/core/gen/banks"
	pbbanksConnect "davensi.com/core/gen/banks/banksconnect"
	pbCommon "davensi.com/core/gen/common"

	"davensi.com/core/internal/util"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName   = "core.banks"
	_banksFields = "id, name, type, bic, bank_code, address_id, contact1_id, contact2_id, contact3_id, " +
		"openbanking_support, parent_id, status"
)

type BankRepository struct {
	pbbanksConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewBankRepository(db *pgxpool.Pool) *BankRepository {
	return &BankRepository{
		db: db,
	}
}

func (s *BankRepository) QbInsert(req *pbBanks.CreateRequest, parentID *uuid.UUID) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleBankValue := []any{}

	// Append required value: login
	qb.SetInsertField("name")
	singleBankValue = append(singleBankValue, req.GetName())
	qb.SetInsertField("type")
	singleBankValue = append(singleBankValue, req.GetType())
	qb.SetInsertField("bic")
	singleBankValue = append(singleBankValue, req.GetBic())
	qb.SetInsertField("bank_code")
	singleBankValue = append(singleBankValue, req.GetBankCode())

	// Append optional fields values
	if req.OpenbankingSupport != nil {
		qb.SetInsertField("openbanking_support")
		singleBankValue = append(singleBankValue, req.GetOpenbankingSupport())
	}
	if req.Status != nil {
		qb.SetInsertField("status")
		singleBankValue = append(singleBankValue, req.GetStatus())
	}
	if req.Parent != nil {
		qb.SetInsertField("parent")
		singleBankValue = append(singleBankValue, parentID)
	}
	// TODO: fill in addressID and the 3 contactIDs

	_, err := qb.SetInsertValues(singleBankValue)

	return qb, err
}

func (s *BankRepository) QbUpdate(
	req *pbBanks.UpdateRequest,
	updateBankID *uuid.UUID,
	_ BankRelationshipIds) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if req.Name != nil {
		qb.SetUpdate("name", req.GetName())
	}

	if req.Type != nil {
		qb.SetUpdate("type", req.GetType())
	}

	if req.Bic != nil {
		qb.SetUpdate("bic", req.GetBic())
	}

	if req.BankCode != nil {
		qb.SetUpdate("bank_code", req.GetBankCode())
	}

	if req.OpenbankingSupport != nil {
		qb.SetUpdate("openbanking_support", req.GetOpenbankingSupport())
	}

	if req.Status != nil {
		qb.SetUpdate("status", req.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", updateBankID.String())

	return qb, nil
}

func (s *BankRepository) QbGetOne(req *pbBanks.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_banksFields)

	switch req.GetSelect().GetSelect().(type) {
	case *pbBanks.Select_ById:
		qb.Where("id = ?", req.GetSelect().GetById())
	case *pbBanks.Select_ByName:
		qb.Where("name = ?", req.GetSelect().GetByName())
	}

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *BankRepository) QbGetList(req *pbBanks.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_banksFields, "banks"))

	if req.Name != nil {
		qb.Where("banks.name LIKE '%' || ? || '%'", req.GetName())
	}

	if req.Type != nil {
		bankTypes := req.GetType().GetList()

		if len(bankTypes) > 0 {
			args := []any{}
			for _, v := range bankTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"banks.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(bankTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if req.Bic != nil {
		qb.Where("banks.bic LIKE '%' || ? || '%'", req.GetBic())
	}

	if req.BankCode != nil {
		qb.Where("banks.bank_code LIKE '%' || ? || '%'", req.GetBankCode())
	}

	if req.OpenbankingSupport != nil {
		qb.Where("banks.openbanking_support = ?", req.GetOpenbankingSupport())
	}

	if req.Status != nil {
		bankStatuses := req.GetStatus().GetList()

		if len(bankStatuses) > 0 {
			args := []any{}
			for _, v := range bankStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"banks.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(bankStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *BankRepository) ScanRow(row pgx.Row) (*pbBanks.Bank, error) {
	var (
		bankID                 pgtype.Text
		bankName               pgtype.Text
		bankType               uint32
		bankBic                pgtype.Text
		bankBankCode           pgtype.Text
		bankAddressID          pgtype.Text
		bankContact1ID         pgtype.Text
		bankContact2ID         pgtype.Text
		bankContact3ID         pgtype.Text
		bankOpenBankingSupport bool
		bankParentID           pgtype.Text
		bankStatus             uint32
	)

	err := row.Scan(
		&bankID,
		&bankName,
		&bankType,
		&bankBic,
		&bankBankCode,
		&bankAddressID,
		&bankContact1ID,
		&bankContact2ID,
		&bankContact3ID,
		&bankOpenBankingSupport,
		&bankParentID,
		&bankStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbBanks.Bank{
		Id:                 bankID.String,
		Name:               bankName.String,
		Type:               pbBanks.Type(bankType),
		Bic:                bankBic.String,
		BankCode:           bankBankCode.String,
		OpenbankingSupport: &bankOpenBankingSupport,
		Status:             pbCommon.Status(bankStatus),
	}, nil
}

func SetQBBySelect(selectBank *pbBanks.Select, qb *util.QueryBuilder) {
	switch selectBank.GetSelect().(type) {
	case *pbBanks.Select_ById:
		qb.Where("banks.id = ?", selectBank.GetById())
	case *pbBanks.Select_ByName:
		qb.Where("banks.name = ?", selectBank.GetByName())
	}
}
