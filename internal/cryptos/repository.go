package cryptos

import (
	"database/sql"
	"errors"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbCryptos "davensi.com/core/gen/cryptos"
	pbUoMs "davensi.com/core/gen/uoms"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.cryptos"
)

type CryptoRepository struct {
	db *pgxpool.Pool
}

func NewCryptoRepository(db *pgxpool.Pool) *CryptoRepository {
	return &CryptoRepository{
		db: db,
	}
}

func (s *CryptoRepository) QbGetOne(_ *pbUoMs.GetRequest, uomsQb *util.QueryBuilder) *util.QueryBuilder {
	uomsQb.Join(fmt.Sprintf("LEFT JOIN %s ON cryptos.id = uoms.id", _tableName))
	uomsQb.Select("cryptos.crypto_type")
	uomsQb.Where("uoms.type = ?", pbUoMs.Type_TYPE_CRYPTO)

	uomsQb.Where("uoms.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return uomsQb
}

func (s *CryptoRepository) QbGetList(msg *pbCryptos.GetListRequest, uomsQb *util.QueryBuilder) *util.QueryBuilder {
	uomsQb.Join(fmt.Sprintf("LEFT JOIN %s ON cryptos.id = uoms.id", _tableName))
	uomsQb.Select("cryptos.crypto_type")
	uomsQb.Where("uoms.type = ?", pbUoMs.Type_TYPE_CRYPTO)

	if msg.CryptoType != nil {
		uomsQb.Where("cryptos.crypto_type = ?", msg.GetCryptoType())
	}

	return uomsQb
}

func (s *CryptoRepository) QbUpdate(msg *pbCryptos.UpdateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableName)

	if msg.CryptoType != nil {
		qb.SetUpdate("crypto_type", msg.GetCryptoType())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetUom().GetSelect().Select.(type) {
	case *pbUoMs.Select_ById:
		qb.Where("cryptos.id = ?", msg.GetUom().GetSelect().GetById())
	case *pbUoMs.Select_ByTypeSymbol:
		qb.Where(
			"cryptos.id = (SELECT uoms.id FROM uoms WHERE uoms.type = ? AND uoms.symbol = ?)",
			msg.GetUom().GetSelect().GetByTypeSymbol().GetType(),
			msg.GetUom().GetSelect().GetByTypeSymbol().GetSymbol(),
		)
	}

	return qb, nil
}

func (s *CryptoRepository) QbInsert(msg *pbCryptos.CreateRequest, uom *pbUoMs.UoM) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("id")
	singleValue = append(singleValue, uom.GetId())

	// Append optional fields values
	if msg.CryptoType != nil {
		qb.SetInsertField("crypto_type")
		singleValue = append(singleValue, msg.GetCryptoType())
	}

	_, err := qb.SetInsertValues(singleValue)

	return qb, err
}

func ScanCryptoUom(row pgx.Row) (*pbCryptos.Crypto, error) {
	var (
		id      string
		uomType pbUoMs.Type
		symbol  string
		// Nullable field must use sql.* type or a scan error will be thrown
		name              sql.NullString
		icon              sql.NullString
		managedDecimals   sql.NullInt32
		displayedDecimals sql.NullInt32
		reportingUnit     sql.NullBool
		status            sql.NullInt32
		cryptoType        sql.NullInt32
	)

	err := row.Scan(
		&id,
		&uomType,
		&symbol,
		&name,
		&icon,
		&managedDecimals,
		&displayedDecimals,
		&reportingUnit,
		&status,
		&cryptoType,
	)
	if err != nil {
		return nil, err
	}

	return &pbCryptos.Crypto{
		Uom: &pbUoMs.UoM{
			Id:                id,
			Type:              uomType,
			Symbol:            symbol,
			Name:              util.GetSQLNullString(name),
			Icon:              util.GetSQLNullString(icon),
			ManagedDecimals:   uint32(util.GetPointInt32(util.GetSQLNullInt32(managedDecimals))),
			DisplayedDecimals: uint32(util.GetPointInt32(util.GetSQLNullInt32(displayedDecimals))),
			ReportingUnit:     util.GetSQLNullBool(reportingUnit),
			Status:            pbCommon.Status(util.GetPointInt32(util.GetSQLNullInt32(status))),
		},
		CryptoType: pbCryptos.Type(util.GetPointInt32(util.GetSQLNullInt32(cryptoType))),
	}, nil
}

func ScanCrypto(row pgx.Row) (*pbCryptos.Crypto, error) {
	var (
		id         string
		cryptoType pbCryptos.Type
	)

	err := row.Scan(
		&id,
		&cryptoType,
	)
	if err != nil {
		return nil, err
	}

	return &pbCryptos.Crypto{
		Uom: &pbUoMs.UoM{
			Id: id,
		},
		CryptoType: cryptoType,
	}, nil
}
