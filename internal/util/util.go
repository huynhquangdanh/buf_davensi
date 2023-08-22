package util

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	pbCommon "davensi.com/core/gen/common"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PgxConn() (*pgxpool.Pool, error) {
	cockroachdbHost := viper.GetString("COCKROACHDB_HOST")
	cockroachdbPort := viper.GetString("COCKROACHDB_PORT")
	cockroachdbUsername := viper.GetString("COCKROACHDB_USERNAME")
	cockroachdbPassword := viper.GetString("COCKROACHDB_PASSWORD")
	cockroachdbDatabase := viper.GetString("COCKROACHDB_DATABASE")
	cockroachdbMaxConn := viper.GetString("COCKROACHDB_MAX_CONN")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?pool_max_conns=%s",
		cockroachdbUsername,
		cockroachdbPassword,
		cockroachdbHost,
		cockroachdbPort,
		cockroachdbDatabase,
		cockroachdbMaxConn,
	)
	dbPool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return dbPool, err
}

// Initialize Database
// func initDatabase(conn *pgx.Conn) {
// crdbpgx.ExecuteTx(context.Background(), conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
// 	files, err := os.ReadDir("sql/")
// 	if err == nil {
// 		for _, f := range files {
// 			var tableName string = strings.TrimSuffix(f.Name(), ".sql")
// 			log.Printf("Checking tables '%s'...", tableName)
// 			content, err := os.ReadFile("sql/" + f.Name())
// 			if err == nil {
// 				if _, err := tx.Exec(context.Background(), string(content)); err != nil {
// 					if strings.Contains(err.Error(), "42P07") {
// 						log.Print("Table already exist.")
// 					} else {
// 						return err
// 					}
// 				} else {
// 					log.Print("Table Created.")
// 				}
// 				log.Println("Ok")
// 			} else {
// 				return err
// 			}
// 		}
// 	} else {
// 		return err
// 	}
// 	return nil
// })
// }

func Contains[T comparable](slice []T, value T) bool {
	for _, n := range slice {
		if value == n {
			return true
		}
	}
	return false
}

func Filter[T comparable](slice []T, fn func(T) bool) []T {
	var result []T

	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}

	return result
}

func Map[T comparable](slice []T, fn func(rowValue T, rowIndex int) T) []T {
	var result []T

	for index, value := range slice {
		result = append(result, fn(value, index))
	}

	return result
}

func GetFieldsWithTableName(fields, tableName string) string {
	return strings.Join(
		Map[string](
			strings.Split(fields, ", "),
			func(rowValue string, _ int) string {
				return fmt.Sprintf("%s.%s", tableName, rowValue)
			},
		),
		", ",
	)
}

func GetSQLNullString(sqlNullString sql.NullString) *string {
	if sqlNullString.Valid {
		return &sqlNullString.String
	}
	return nil
}

func GetPointString(pointString *string) string {
	if pointString == nil {
		return ""
	}
	return *pointString
}

func GetPointInt32(pointInt32 *int32) int32 {
	if pointInt32 == nil {
		return 0
	}
	return *pointInt32
}

func GetSQLNullBool(sqlBool sql.NullBool) bool {
	if sqlBool.Valid {
		return sqlBool.Bool
	}
	return false
}

func GetSQLNullTime(sqlNullTimme sql.NullTime) *timestamppb.Timestamp {
	if sqlNullTimme.Valid {
		return timestamppb.New(sqlNullTimme.Time)
	}
	return nil
}

func GetSQLNullFloat(sqlNullFloat sql.NullFloat64) *float64 {
	if sqlNullFloat.Valid {
		return &sqlNullFloat.Float64
	}
	return nil
}

func GetSQLNullInt16(sqlNullInt16 sql.NullInt16) *int16 {
	if sqlNullInt16.Valid {
		return &sqlNullInt16.Int16
	}
	return nil
}

func GetSQLNullInt32(sqlNullFloat sql.NullInt32) *int32 {
	if sqlNullFloat.Valid {
		return &sqlNullFloat.Int32
	}
	return nil
}

func GetDBTimestampValue(time *timestamppb.Timestamp) string {
	return time.AsTime().Format("2006-01-02 15:04:05")
}

func FloatToDecimal(number *float64) *pbCommon.Decimal {
	if number != nil {
		return &pbCommon.Decimal{
			Value: fmt.Sprintf("%f", *number),
		}
	}
	return nil
}

// Map manipulates a slice and transforms it to a slice of another type.
// Play: https://go.dev/play/p/OkPcYAhBo0D
func MapTToR[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item, i)
	}

	return result
}

// ForEach iterates over elements of collection and invokes iteratee for each element.
// Play: https://go.dev/play/p/oofyiUPRf8t
func ForEach[T any](collection []T, iteratee func(item T, index int)) {
	for i, item := range collection {
		iteratee(item, i)
	}
}
