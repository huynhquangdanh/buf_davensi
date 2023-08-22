package useraddresses

import (
	"context"
	"sync"

	pbAddresses "davensi.com/core/gen/addresses"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type ServiceServer struct {
	repo UserAddressesRepository
	db   *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewUserAddressRepository(db),
		db:   db,
	}
}

var (
	singleSS *ServiceServer
	onceSS   sync.Once
)

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	onceSS.Do(func() {
		singleSS = NewServiceServer(db)
	})
	return singleSS
}

func (s *ServiceServer) UpsertUserAddresses(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	addresses []*pbAddresses.SetLabeledAddress,
	command string,
) (result []*pbAddresses.LabeledAddress, queryErr error) {
	result = []*pbAddresses.LabeledAddress{}
	qbUserAddress, err := s.repo.QbUpsertUserAddresses(userID, addresses, command)
	if err != nil {
		return result, err
	}
	var rowsUpsertUserAddress pgx.Rows

	sqlUpsertUserAddress, sqlUpsertUserAddressArgs, _ := qbUserAddress.GenerateSQL()
	log.Info().Msgf("query upsert user address %s with args: %s", sqlUpsertUserAddress, sqlUpsertUserAddressArgs)
	if tx != nil {
		rowsUpsertUserAddress, queryErr = tx.Query(ctx, sqlUpsertUserAddress, sqlUpsertUserAddressArgs...)
		if err != nil {
			return result, queryErr
		}
	} else {
		rowsUpsertUserAddress, queryErr = s.db.Query(ctx, sqlUpsertUserAddress, sqlUpsertUserAddressArgs...)
		if err != nil {
			return result, queryErr
		}
	}
	defer rowsUpsertUserAddress.Close()
	result, queryErr = s.repo.ScanRow(rowsUpsertUserAddress)
	return result, queryErr
}
