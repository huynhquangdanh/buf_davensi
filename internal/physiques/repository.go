package physiques

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbKyc "davensi.com/core/gen/kyc"
	pbPhysiques "davensi.com/core/gen/physiques"
	pbPhysiquesConnect "davensi.com/core/gen/physiques/physiquesconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PhysiqueRepository struct {
	pbPhysiquesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewPhysiqueRepository(db *pgxpool.Pool) *PhysiqueRepository {
	return &PhysiqueRepository{
		db: db,
	}
}

func (s *PhysiqueRepository) QbInsert(msg *pbPhysiques.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singlePhysiqueValue := []any{}

	// Append optional fields values
	if msg.Race != nil {
		qb.SetInsertField("race")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetRace())
	}
	if msg.Ethnicity != nil {
		qb.SetInsertField("ethnicity")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetEthnicity())
	}
	if msg.BodyShape != nil {
		qb.SetInsertField("body_shape")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetBodyShape())
	}
	if msg.Height != nil {
		qb.SetInsertField("height")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetHeight())
	}
	if msg.Weight != nil {
		qb.SetInsertField("weight")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetWeight())
	}
	if msg.EyesColor != nil {
		qb.SetInsertField("eyes_color")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetEyesColor())
	}
	if msg.HairColor != nil {
		qb.SetInsertField("hair_color")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetHairColor())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singlePhysiqueValue = append(singlePhysiqueValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singlePhysiqueValue)

	return qb, err
}

func (s *PhysiqueRepository) QbUpdate(msg *pbPhysiques.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Race != nil {
		qb.SetUpdate("race", msg.GetRace())
	}
	if msg.Ethnicity != nil {
		qb.SetUpdate("ethnicity", msg.GetEthnicity())
	}
	if msg.BodyShape != nil {
		qb.SetUpdate("body_shape", msg.GetBodyShape())
	}
	if msg.Height != nil {
		qb.SetUpdate("height", msg.GetHeight())
	}
	if msg.Weight != nil {
		qb.SetUpdate("weight", msg.GetWeight())
	}
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}
	if msg.EyesColor != nil {
		qb.SetUpdate("eyes_color", msg.GetEyesColor())
	}
	if msg.HairColor != nil {
		qb.SetUpdate("hair_color", msg.GetHairColor())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", msg.GetId())

	return qb, nil
}

func (s *PhysiqueRepository) QbGetOne(msg *pbPhysiques.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	qb.Where("id = ? AND status = ?", msg.GetId(), pbKyc.Status_STATUS_VALIDATED)

	return qb
}

func (s *PhysiqueRepository) QbGetList(msg *pbPhysiques.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.Race != nil {
		qb.Where("race LIKE '%' || ? || '%'", msg.GetRace())
	}
	if msg.Ethnicity != nil {
		qb.Where("ethnicity LIKE '%' || ? || '%'", msg.GetEthnicity())
	}
	if msg.BodyShape != nil {
		qb.Where("body_shape LIKE '%' || ? || '%'", msg.GetBodyShape())
	}
	if msg.Height != nil {
		qb.Where("height LIKE '%' || ? || '%'", msg.GetHeight())
	}
	if msg.Weight != nil {
		qb.Where("weight LIKE '%' || ? || '%'", msg.GetWeight())
	}
	if msg.EyesColor != nil {
		qb.Where("eyes_color LIKE '%' || ? || '%'", msg.GetEyesColor())
	}
	if msg.HairColor != nil {
		qb.Where("hair_color LIKE '%' || ? || '%'", msg.GetHairColor())
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

func (s *PhysiqueRepository) ScanRow(row pgx.Row) (*pbPhysiques.Physique, error) {
	// TODO: add new 2 fields from proto which isn't existing DB
	var (
		id                string
		nullableRace      sql.NullString
		race              *string
		nullableEthnicity sql.NullString
		ethnicity         *string
		nullableBodyShape sql.NullString
		bodyShape         *string
		nullableHeight    sql.NullString
		height            *string
		nullableWeight    sql.NullString
		weight            *string
		nullableEyesColor sql.NullString
		eyesColor         *string
		nullableHairColor sql.NullString
		hairColor         *string
		status            pbKyc.Status
	)

	err := row.Scan(
		&id,
		&nullableRace,
		&nullableEthnicity,
		&nullableBodyShape,
		&nullableHeight,
		&nullableWeight,
		&status,
		&nullableEyesColor,
		&nullableHairColor,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing
	if nullableRace.Valid {
		race = &nullableRace.String
	}
	if nullableEthnicity.Valid {
		ethnicity = &nullableEthnicity.String
	}
	if nullableBodyShape.Valid {
		bodyShape = &nullableBodyShape.String
	}
	if nullableHeight.Valid {
		height = &nullableHeight.String
	}
	if nullableWeight.Valid {
		weight = &nullableWeight.String
	}
	if nullableEyesColor.Valid {
		eyesColor = &nullableEyesColor.String
	}
	if nullableHairColor.Valid {
		hairColor = &nullableHairColor.String
	}

	return &pbPhysiques.Physique{
		Id:        id,
		Race:      race,
		Ethnicity: ethnicity,
		EyesColor: eyesColor,
		HairColor: hairColor,
		BodyShape: bodyShape,
		Height:    height,
		Weight:    weight,
		Status:    status,
	}, nil
}
