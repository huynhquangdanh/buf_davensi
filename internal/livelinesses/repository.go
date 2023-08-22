package livelinesses

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbKyc "davensi.com/core/gen/kyc"
	pbLiveliveness "davensi.com/core/gen/liveliness"
	pbLivelinessConnect "davensi.com/core/gen/liveliness/livelinessconnect"
	"davensi.com/core/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LivelinessRepository struct {
	pbLivelinessConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewLivelinessRepository(db *pgxpool.Pool) *LivelinessRepository {
	return &LivelinessRepository{
		db: db,
	}
}

func (s *LivelinessRepository) QbInsert(msg *pbLiveliveness.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _table)
	singleLivelinessValue := []any{}

	// Append optional fields values
	if msg.LivelinessVideoFile != nil {
		qb.SetInsertField("liveliness_video_file")
		singleLivelinessValue = append(singleLivelinessValue, msg.GetLivelinessVideoFile())
	}
	if msg.LivelinessVideoFileType != nil {
		qb.SetInsertField("liveliness_video_file_type")
		singleLivelinessValue = append(singleLivelinessValue, msg.GetLivelinessVideoFileType())
	}
	if msg.TimestampVideoFile != nil {
		qb.SetInsertField("timestamp_video_file")
		singleLivelinessValue = append(singleLivelinessValue, msg.GetTimestampVideoFile())
	}
	if msg.TimestampVideoFileType != nil {
		qb.SetInsertField("timestamp_video_file_type")
		singleLivelinessValue = append(singleLivelinessValue, msg.GetTimestampVideoFileType())
	}
	if msg.IdOwnershipPhotoFile != nil {
		qb.SetInsertField("id_ownership_photo_file")
		singleLivelinessValue = append(singleLivelinessValue, msg.GetIdOwnershipPhotoFile())
	}

	if msg.IdOwnershipPhotoFileType != nil {
		qb.SetInsertField("id_ownership_photo_file_type")
		singleLivelinessValue = append(singleLivelinessValue, msg.GetIdOwnershipPhotoFileType())
	}

	if msg.Status != nil {
		qb.SetInsertField("status")
		singleLivelinessValue = append(singleLivelinessValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleLivelinessValue)

	return qb, err
}

func (s *LivelinessRepository) QbUpdate(msg *pbLiveliveness.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _table)

	if msg.LivelinessVideoFile != nil {
		qb.SetUpdate("liveliness_video_file", msg.GetLivelinessVideoFile())
	}

	if msg.LivelinessVideoFileType != nil {
		qb.SetUpdate("liveliness_video_file_type", msg.GetLivelinessVideoFileType())
	}

	if msg.TimestampVideoFile != nil {
		qb.SetUpdate("timestamp_video_file", msg.GetTimestampVideoFile())
	}

	if msg.TimestampVideoFileType != nil {
		qb.SetUpdate("timestamp_video_file_type", msg.GetTimestampVideoFileType())
	}

	if msg.IdOwnershipPhotoFile != nil {
		qb.SetUpdate("id_ownership_photo_file", msg.GetIdOwnershipPhotoFile())
	}

	if msg.IdOwnershipPhotoFileType != nil {
		qb.SetUpdate("id_ownership_photo_file_type", msg.GetIdOwnershipPhotoFileType())
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", msg.GetId())

	return qb, nil
}

func (s *LivelinessRepository) QbGetOne(msg *pbLiveliveness.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _table)
	qb.Select(_livelinessFields)

	qb.Where("id = ?", msg.GetId())

	return qb
}

func (s *LivelinessRepository) QbGetList(msg *pbLiveliveness.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _table)
	qb.Select(_livelinessFields)

	if msg.LivelinessVideoFile != nil {
		qb.Where("liveliness_video_file LIKE '%' || ? || '%'", msg.GetIdOwnershipPhotoFile())
	}
	if msg.LivelinessVideoFileType != nil {
		qb.Where("liveliness_video_file_type LIKE '%' || ? || '%'", msg.GetLivelinessVideoFileType())
	}
	if msg.TimestampVideoFile != nil {
		qb.Where("timestamp_video_file LIKE '%' || ? || '%'", msg.GetTimestampVideoFile())
	}
	if msg.TimestampVideoFileType != nil {
		qb.Where("timestamp_video_file_type LIKE '%' || ? || '%'", msg.GetTimestampVideoFileType())
	}

	if msg.IdOwnershipPhotoFile != nil {
		qb.Where("id_ownership_photo_file LIKE '%' || ? || '%'", msg.GetIdOwnershipPhotoFile())
	}

	if msg.IdOwnershipPhotoFileType != nil {
		qb.Where("id_ownership_photo_file_type LIKE '%' || ? || '%'", msg.GetIdOwnershipPhotoFileType())
	}

	if msg.Status != nil {
		livelinessStatus := msg.GetStatus().GetList()

		if len(livelinessStatus) > 0 {
			args := []any{}
			for _, v := range livelinessStatus {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"status IN (%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(livelinessStatus)), ""), ", "),
				),
				args...,
			)
		}
	} else {
		qb.Where("status = ?", pbKyc.Status_STATUS_VALIDATED)
	}

	return qb
}

func (s *LivelinessRepository) ScanRow(row pgx.Row) (*pbLiveliveness.Liveliness, error) {
	// TODO: add new 2 fields from proto which isn't existing DB
	var (
		id                       string
		livelinessVideoFile      sql.NullString
		livelinessVideoFileType  sql.NullString
		timestampVideoFile       sql.NullString
		timestampVideoFileType   sql.NullString
		idOwnershipPhotoFile     sql.NullString
		idOwnershipPhotoFileType sql.NullString
		status                   pbKyc.Status
	)

	err := row.Scan(
		&id,
		&livelinessVideoFile,
		&livelinessVideoFileType,
		&timestampVideoFile,
		&timestampVideoFileType,
		&idOwnershipPhotoFile,
		&idOwnershipPhotoFileType,
		&status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing

	return &pbLiveliveness.Liveliness{
		Id:                       id,
		LivelinessVideoFile:      util.GetSQLNullString(livelinessVideoFile),
		LivelinessVideoFileType:  util.GetSQLNullString(livelinessVideoFileType),
		TimestampVideoFile:       util.GetSQLNullString(timestampVideoFile),
		TimestampVideoFileType:   util.GetSQLNullString(timestampVideoFileType),
		IdOwnershipPhotoFile:     util.GetSQLNullString(idOwnershipPhotoFile),
		IdOwnershipPhotoFileType: util.GetSQLNullString(idOwnershipPhotoFileType),
		Status:                   status,
	}, nil
}
