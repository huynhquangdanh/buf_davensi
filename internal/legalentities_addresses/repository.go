package legalentitiesaddresses

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	pbAddresses "davensi.com/core/gen/addresses"
	pbKyc "davensi.com/core/gen/kyc"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.legalentities_addresses"
	_fields    = "legalentity_id, label, address_id, status"
)

type LegalEntityAddressesRepository struct {
	db *pgxpool.Pool
}

var (
	singleRepo *LegalEntityAddressesRepository
	once       sync.Once
)

func NewLegalEntityAddressRepository(db *pgxpool.Pool) *LegalEntityAddressesRepository {
	return &LegalEntityAddressesRepository{
		db: db,
	}
}

func GetSingletonRepository(db *pgxpool.Pool) *LegalEntityAddressesRepository {
	once.Do(func() {
		singleRepo = NewLegalEntityAddressRepository(db)
	})
	return singleRepo
}

func (r *LegalEntityAddressesRepository) QbUpsertUserAddresses(
	legalEntityID string,
	addresses []*pbAddresses.SetLabeledAddress,
	command string) (*util.QueryBuilder, error) {
	var queryType util.QueryType
	if command == "upsert" {
		queryType = util.Upsert
	} else {
		queryType = util.Insert
	}
	qb := util.CreateQueryBuilder(queryType, _tableName)
	qb.SetInsertField("legalentity_id", "label", "address_id", "status")

	for _, address := range addresses {
		_, err := qb.SetInsertValues([]any{
			legalEntityID,
			address.GetLabel(),
			address.GetId(),
			// address.GetMainAddress(),
			// address.GetOwnershipStatus().Enum(),
			address.GetStatus().Enum(),
		})
		if err != nil {
			return nil, err
		}
	}

	return qb, nil
}

func (r *LegalEntityAddressesRepository) ScanRow(row pgx.Row) (*pbAddresses.LabeledAddress, error) {
	var (
		legalEntityID string
		label         string
		addressID     string
		status        *pbKyc.Status
	)
	err := row.Scan(&legalEntityID, &label, &addressID, &status)
	if err != nil {
		return nil, err
	}
	return &pbAddresses.LabeledAddress{
		Label: label,
		Address: &pbAddresses.Address{
			Id: addressID,
		},
		Status: *status.Enum(),
	}, nil
}

func (r *LegalEntityAddressesRepository) QbGetOne(leID, label string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	qb.Where("legalentity_id = ? AND label = ?", leID, label)

	return qb
}

func (r *LegalEntityAddressesRepository) QbUpdateLegalEntitiesAddresses(
	legalEntityID string,
	req *pbAddresses.UpdateLabeledAddressRequest,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableName)

	if req.Status != nil {
		qb.SetUpdate("status", req.GetStatus()).
			Where("label = ? AND legalentity_id = ?", req.GetByLabel(), legalEntityID)
		return qb, nil
	}

	return nil, errors.New("cannot update without new value")
}

func (r *LegalEntityAddressesRepository) QbRemoveLegalEntitiesAddresses(
	legalEntityID string,
	req *pbLegalEntities.LabelList,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetUpdate("status", pbKyc.Status_STATUS_CANCELED)
	if !qb.IsUpdatable() {
		return qb, errors.New("cannot remove")
	}

	labels := req.GetList()

	args := []any{}
	for _, v := range labels {
		args = append(args, v)
	}

	qb.Where(
		fmt.Sprintf(
			"label IN(%s)",
			strings.Join(strings.Split(strings.Repeat("?", len(req.GetList())), ""), ", "),
		),
		args...,
	)

	qb.Where("legalentity_id = ?", legalEntityID)

	return qb, nil
}
