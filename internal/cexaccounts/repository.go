package cexaccounts

import (
	"errors"

	pbCexAccount "davensi.com/core/gen/cexaccounts"
	pbFSProvider "davensi.com/core/gen/fsproviders"
	pbRecipient "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CexAccountRepository struct {
	db *pgxpool.Pool
}

func NewCexAccountRepository(db *pgxpool.Pool) *CexAccountRepository {
	return &CexAccountRepository{
		db: db,
	}
}

//nolint:all
func (r *CexAccountRepository) GetInsertOrUpdateFields(message any) ([]any, []string, []any) {
	switch message.(type) {
	case *pbCexAccount.CreateRequest:
		msg := message.(*pbCexAccount.CreateRequest)
		return []any{msg.Provider}, []string{"fs_provider"}, []any{msg.GetProvider()}
	case *pbCexAccount.UpdateRequest:
		msg := message.(*pbCexAccount.UpdateRequest)
		return []any{msg.Provider}, []string{"fs_provider"}, []any{msg.GetProvider()}
	default:
		return []any{}, []string{}, []any{}
	}
}

func (r *CexAccountRepository) QbInsert(recipientID, providerID string) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _table)
	var singleCexAccountValue []any
	qb.SetInsertField("id")
	singleCexAccountValue = append(singleCexAccountValue, recipientID)
	qb.SetInsertField("fsprovider_id")
	singleCexAccountValue = append(singleCexAccountValue, providerID)
	_, err := qb.SetInsertValues(singleCexAccountValue)
	qb.SetReturnFields("*")
	return qb, err
}

func (r *CexAccountRepository) QbUpdate(recipientID, providerID string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _table)
	qb.SetUpdate("fsprovider_id", providerID)
	err = nil
	if !qb.IsUpdatable() {
		err = errors.New("cannot update without new value")
		return
	}
	qb.Where("id = ?", recipientID)
	return
}

func (s *CexAccountRepository) QbGetOne(recipientID string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _table)
	qb.Where("id = ?", recipientID)
	qb.Select(_fields)

	return qb
}

func (r *CexAccountRepository) QbDelete(recipientID string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Delete, _table)
	qb.Where("id = ?", recipientID)
	return qb
}

func (r *CexAccountRepository) ScanRow(row pgx.Row) (*pbCexAccount.CExAccount, error) {
	var (
		id           string
		fsProviderID string
	)

	err := row.Scan(
		&id,
		&fsProviderID,
	)
	if err != nil {
		return nil, err
	}

	return &pbCexAccount.CExAccount{
		Recipient: &pbRecipient.Recipient{
			Id: id,
		},
		Provider: &pbFSProvider.FSProvider{
			Id: fsProviderID,
		},
	}, nil
}
