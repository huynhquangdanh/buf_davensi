package credentials

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	pbCountries "davensi.com/core/gen/countries"
	pbCredential "davensi.com/core/gen/credentials"
	pbCredentialConnect "davensi.com/core/gen/credentials/credentialsconnect"
	pbKyc "davensi.com/core/gen/kyc"
	"davensi.com/core/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CredentialRepository struct {
	pbCredentialConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewCredentialRepository(db *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{
		db: db,
	}
}

//nolint:all
func (r *CredentialRepository) GetInsertOrUpdateFields(message any, command string) ([]any, []string, []any) {
	if command == "CREATE" {
		if msg, ok := message.(*pbCredential.CreateRequest); ok {
			return []any{
					msg.Photo, msg.Gender, msg.Title, msg.FirstName,
					msg.MiddleNames, msg.LastName, msg.Birthday, msg.CountryOfBirth,
					msg.CountryOfNationality, msg.Status,
				},
				[]string{
					"photo", "gender", "title", "first_name", "middle_names", "last_name", "birthday",
					"country_of_birth_id", "country_of_nationality_id", "status",
				},
				[]any{
					msg.GetPhoto(), msg.GetGender(), msg.GetTitle(),
					msg.GetFirstName(), msg.GetMiddleNames(), msg.GetLastName(), util.GetDBTimestampValue(msg.GetBirthday()),
					msg.GetCountryOfBirth().GetById(), msg.GetCountryOfNationality().GetById(), msg.GetStatus(),
				}
		} else {
			return []any{}, []string{}, []any{}
		}
	} else if command == "UPDATE" {
		if msg, ok := message.(*pbCredential.UpdateRequest); ok {
			return []any{
					msg.Photo, msg.Gender, msg.Title, msg.FirstName,
					msg.MiddleNames, msg.LastName, msg.Birthday, msg.CountryOfBirth,
					msg.CountryOfNationality, msg.Status,
				},
				[]string{
					"photo", "gender", "title", "first_name", "middle_names", "last_name", "birthday",
					"country_of_birth_id", "country_of_nationality_id", "status",
				},
				[]any{
					msg.GetPhoto(), msg.GetGender(), msg.GetTitle(),
					msg.GetFirstName(), msg.GetMiddleNames(), msg.GetLastName(), util.GetDBTimestampValue(msg.GetBirthday()),
					msg.GetCountryOfBirth().GetById(), msg.GetCountryOfNationality().GetById(), msg.GetStatus(),
				}
		} else {
			return []any{}, []string{}, []any{}
		}
	}
	return []any{}, []string{}, []any{}
}

func (r *CredentialRepository) QbInsert(msg *pbCredential.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _table)

	singleCredentialValue := []any{}
	// Append optional fields values
	fields, titles, getFields := r.GetInsertOrUpdateFields(msg, "CREATE")
	for i := 0; i < len(fields); i++ {
		field, title, getField := fields[i], titles[i], getFields[i]
		if !reflect.ValueOf(field).IsNil() {
			singleCredentialValue = append(singleCredentialValue, getField)
			qb.SetInsertField(title)
		}
	}

	_, err := qb.SetInsertValues(singleCredentialValue)
	return qb, err
}

func (r *CredentialRepository) QbUpdate(msg *pbCredential.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _table)

	fields, titles, getFields := r.GetInsertOrUpdateFields(msg, "UPDATE")
	for i := 0; i < len(fields); i++ {
		field, title, getField := fields[i], titles[i], getFields[i]
		if !reflect.ValueOf(field).IsNil() {
			qb.SetUpdate(title, getField)
		}
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", msg.GetId())

	return qb, nil
}

func (s *CredentialRepository) QbGetOne(msg *pbCredential.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _table)
	qb.Select(_fields)

	qb.Where("id = ?", msg.GetId())
	qb.Where("status = ?", pbKyc.Status_STATUS_VALIDATED)

	return qb
}

func (s *CredentialRepository) QbGetList(msg *pbCredential.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _table)
	qb.Select(_fields)

	if msg.Photo != nil {
		qb.Where("photo LIKE '%' || ? || '%'", msg.GetPhoto())
	}
	if msg.Gender != nil {
		qb.Where("gender LIKE '%' || ? || '%'", msg.GetGender())
	}
	if msg.Title != nil {
		qb.Where("title LIKE '%' || ? || '%'", msg.GetTitle())
	}
	if msg.FirstName != nil {
		qb.Where("first_name LIKE '%' || ? || '%'", msg.GetFirstName())
	}

	if msg.MiddleNames != nil {
		qb.Where("middle_names LIKE '%' || ? || '%'", msg.GetMiddleNames())
	}

	if msg.LastName != nil {
		qb.Where("last_name LIKE '%' || ? || '%'", msg.GetLastName())
	}

	if msg.Birthday != nil {
		qb.Where("birthday = ?", msg.GetBirthday())
	}

	// for status
	if msg.Status != nil {
		credentialStatuses := msg.GetStatus().GetList()

		if len(credentialStatuses) > 0 {
			args := []any{}
			for _, v := range credentialStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"status IN (%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(credentialStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *CredentialRepository) ScanRow(row pgx.Row) (*pbCredential.Credentials, error) {
	// TODO: add new 2 fields from proto which isn't existing DB
	var (
		id                     string
		photo                  sql.NullString
		gender                 sql.NullString
		title                  sql.NullString
		firstName              sql.NullString
		middleNames            sql.NullString
		lastName               sql.NullString
		birthday               sql.NullTime
		countryOfBirthID       sql.NullString
		countryOfNationalityID sql.NullString
		status                 pbKyc.Status
	)

	err := row.Scan(
		&id,
		&photo,
		&gender,
		&title,
		&firstName,
		&middleNames,
		&lastName,
		&birthday,
		&countryOfBirthID,
		&countryOfNationalityID,
		&status,
	)

	if err != nil {
		return nil, err
	}
	// Nullable field processing

	return &pbCredential.Credentials{
		Id:                   id,
		Photo:                util.GetSQLNullString(photo),
		Gender:               util.GetSQLNullString(gender),
		Title:                util.GetSQLNullString(title),
		FirstName:            util.GetSQLNullString(firstName),
		MiddleNames:          util.GetSQLNullString(middleNames),
		LastName:             util.GetSQLNullString(lastName),
		Birthday:             util.GetSQLNullTime(birthday),
		CountryOfBirth:       &pbCountries.Country{Id: getNullableString(countryOfBirthID)},
		CountryOfNationality: &pbCountries.Country{Id: getNullableString(countryOfNationalityID)},
		Status:               status,
	}, nil
}
func getNullableString(s sql.NullString) string {
	if s.Valid {
		return *util.GetSQLNullString(s)
	}
	return *util.GetSQLNullString(sql.NullString{
		String: "",
		Valid:  true,
	})
}
