package common

import (
	"context"
	"errors"

	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type QueryStatements struct {
	SQLStr  string
	SQLArgs []any
}

func TxWrite[T any](
	ctx context.Context,
	tx pgx.Tx,
	sqlStr string,
	args []any,
	scanRow func(_rows pgx.Row) (*T, error),
) (returnable *T, err error) {
	rows, err := tx.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rowErr := rows.Err()

	if rowErr != nil {
		return nil, rowErr
	}
	if !rows.Next() {
		return nil, errors.New("write 0 record to database")
	}

	return scanRow(rows)
}

// Will remove after change all ScanRow to ScanRows
func TxWriteMulti[T any](
	ctx context.Context,
	tx pgx.Tx,
	sqlStr string,
	args []any,
	scanRow func(_rows pgx.Rows) (*T, error),
) (returnable *T, err error) {
	rows, err := tx.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rowErr := rows.Err()

	if rowErr != nil {
		return nil, rowErr
	}

	return scanRow(rows)
}

func ExecuteTxWrite[T any](
	ctx context.Context,
	conn *pgxpool.Pool,
	sqlStr string,
	args []any,
	scanRow func(_rows pgx.Row) (*T, error),
) (returnable *T, err error) {
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		rowScanned, scanErr := TxWrite[T](
			ctx,
			tx,
			sqlStr,
			args,
			scanRow,
		)

		returnable = rowScanned

		return scanErr
	}); err != nil {
		return nil, err
	}
	return returnable, nil
}

func TxBulkWrite[T any](
	ctx context.Context,
	tx pgx.Tx,
	sqlStr string,
	args []any,
	scanRows func(_rows pgx.Rows) ([]*T, error),
) (returnable []*T, err error) {
	rows, err := tx.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := scanRows(rows)
	return result, err
}

func TxSendBatch[T any](
	ctx context.Context,
	tx pgx.Tx,
	batch *pgx.Batch,
	queryCount int,
	scanRows func(_rows pgx.Rows) (*T, error),
	scanRow func(_row pgx.Row) (*T, error),
	useMultipleRowsScanner bool,
) (returnable []*T, brErr error) {
	results := tx.SendBatch(ctx, batch)
	defer func(results pgx.BatchResults) {
		err := results.Close()
		if err != nil {
			log.Error().Err(err).Msg("close results")
		}
	}(results)
	for i := 0; i < queryCount; i++ {
		rows, brqErr := results.Query()
		if brqErr != nil {
			return nil, brqErr
		}
		var (
			returnableElem *T
			scanErr        error
		)
		for rows.Next() {
			if useMultipleRowsScanner {
				returnableElem, scanErr = scanRows(rows)
			} else {
				returnableElem, scanErr = scanRow(rows)
			}
			if scanErr != nil {
				return nil, scanErr
			}
			returnable = append(returnable, returnableElem)
		}

		if rows.Err() != nil {
			return nil, brErr
		}
	}

	return returnable, nil
}

func TxMultipleStatement[T any](
	ctx context.Context,
	tx pgx.Tx,
	statements []*QueryStatements,
	scanRows func(_row pgx.Rows) ([]*T, error),
) (returnable []*T, brErr error) {
	for _, statement := range statements {
		data, err := TxBulkWrite[T](ctx, tx, statement.SQLStr, statement.SQLArgs, scanRows)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("data from TxWrite %v", data)
		returnable = append(returnable, data...)
	}
	return returnable, nil
}
