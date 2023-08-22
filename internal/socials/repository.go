package socials

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbKyc "davensi.com/core/gen/kyc"
	pbSocials "davensi.com/core/gen/socials"
	pbSocialsConnect "davensi.com/core/gen/socials/socialsconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.kyc_socials"
	_fields    = "id, relationship_status, religion, social_class, profession, status"
)

type SocialRepository struct {
	pbSocialsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewSocialRepository(db *pgxpool.Pool) *SocialRepository {
	return &SocialRepository{
		db: db,
	}
}

func (s *SocialRepository) QbInsert(id string, msg *pbSocials.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleSocialValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("id")
	singleSocialValue = append(singleSocialValue, id)

	// Append optional fields values
	if msg.RelationshipStatus != nil {
		qb.SetInsertField("relationship_status")
		singleSocialValue = append(singleSocialValue, msg.GetRelationshipStatus())
	}
	if msg.Religion != nil {
		qb.SetInsertField("religion")
		singleSocialValue = append(singleSocialValue, msg.GetReligion())
	}
	if msg.SocialClass != nil {
		qb.SetInsertField("social_class")
		singleSocialValue = append(singleSocialValue, msg.GetSocialClass())
	}
	if msg.Profession != nil {
		qb.SetInsertField("profession")
		singleSocialValue = append(singleSocialValue, msg.GetProfession())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleSocialValue = append(singleSocialValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleSocialValue)

	return qb, err
}

func (s *SocialRepository) QbUpdate(msg *pbSocials.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)
	log.Info().Msgf("Update: %v", msg)
	if msg.RelationshipStatus != nil {
		qb.SetUpdate("relationship_status", msg.GetRelationshipStatus())
	}

	if msg.Religion != nil {
		qb.SetUpdate("religion", msg.GetReligion())
	}

	if msg.SocialClass != nil {
		qb.SetUpdate("social_class", msg.GetSocialClass())
	}

	if msg.Profession != nil {
		qb.SetUpdate("profession", msg.GetProfession())
	}

	if msg.Status != nil {
		log.Info().Msg("status")
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", msg.GetId())

	return qb, nil
}

func (s *SocialRepository) QbGetOne(msg *pbSocials.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	qb.Where("id = ? AND status != ?", msg.GetId(), pbKyc.Status_STATUS_CANCELED)

	return qb
}

func (s *SocialRepository) QbGetList(msg *pbSocials.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.RelationshipStatus != nil {
		qb.Where("relationship_status LIKE '%' || ? || '%'", msg.GetRelationshipStatus())
	}
	if msg.Religion != nil {
		qb.Where("religion LIKE '%' || ? || '%'", msg.GetReligion())
	}
	if msg.SocialClass != nil {
		qb.Where("social_class LIKE '%' || ? || '%'", msg.GetSocialClass())
	}
	if msg.Profession != nil {
		qb.Where("profession LIKE '%' || ? || '%'", msg.GetProfession())
	}
	if msg.Status != nil {
		socialStatuses := msg.GetStatus().GetList()

		if len(socialStatuses) > 0 {
			args := []any{}
			for _, v := range socialStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(socialStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func ScanRow(row pgx.Row) (*pbSocials.Social, error) {
	// TODO: add new 2 fields from proto which isn't existing DB
	var (
		id                         string
		nullableRelationshipStatus sql.NullString
		nullableReligion           sql.NullString
		nullableSocialClass        sql.NullString
		nullableProfession         sql.NullString
		status                     pbCommon.Status
	)

	err := row.Scan(
		&id,
		&nullableRelationshipStatus,
		&nullableReligion,
		&nullableSocialClass,
		&nullableProfession,
		&status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing
	var relationshipStatus *string = nil
	if nullableRelationshipStatus.Valid {
		relationshipStatus = &nullableRelationshipStatus.String
	}
	var religion *string = nil
	if nullableReligion.Valid {
		religion = &nullableReligion.String
	}
	var socialClass *string = nil
	if nullableSocialClass.Valid {
		socialClass = &nullableSocialClass.String
	}
	var profession *string = nil
	if nullableProfession.Valid {
		profession = &nullableProfession.String
	}

	return &pbSocials.Social{
		Id:                 id,
		RelationshipStatus: relationshipStatus,
		Religion:           religion,
		SocialClass:        socialClass,
		Profession:         profession,
		Status:             pbKyc.Status(status),
	}, nil
}
