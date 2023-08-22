package defiwallets

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbBlockchains "davensi.com/core/gen/blockchains"
	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbDataSourcesConnect "davensi.com/core/gen/datasources/datasourcesconnect"
	pbDefiwallets "davensi.com/core/gen/defiwallets"
	pbFsproviders "davensi.com/core/gen/fsproviders"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbOrgs "davensi.com/core/gen/orgs"
	pbRecipients "davensi.com/core/gen/recipients"
	pbUsers "davensi.com/core/gen/users"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_table  = "core.defiwallets"
	_fields = "id, blockchain_id, address"
)

type DefiWalletRepository struct {
	pbDataSourcesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewDefiWalletRepository(db *pgxpool.Pool) *DefiWalletRepository {
	return &DefiWalletRepository{
		db: db,
	}
}

func (s *DefiWalletRepository) QbInsert(msg *pbDefiwallets.CreateRequest, recipient *pbRecipients.Recipient) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _table)
	singleFSProviderValue := []any{}

	qb.SetInsertField("id")
	singleFSProviderValue = append(singleFSProviderValue, recipient.GetId())

	// Append optional fields values
	if msg.Address != "" {
		qb.SetInsertField("address")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetAddress())
	}
	if msg.Blockchain != nil {
		qb.SetInsertField("blockchain_id")
		singleFSProviderValue = append(singleFSProviderValue, msg.GetBlockchain().GetById())
	}

	_, err := qb.SetInsertValues(singleFSProviderValue)

	return qb, err
}

func (s *DefiWalletRepository) ScanUpdateRow(row pgx.Row) (*pbDefiwallets.DeFiWallet, error) {
	var (
		botID             sql.NullString
		botType           sql.NullInt32
		defaultParamsName sql.NullString
		botStateValue     sql.NullInt32
		_status           sql.NullInt32
	)

	err := row.Scan(
		&botID,
		&botType,
		&defaultParamsName,
		&botStateValue,
		&_status,
	)
	if err != nil {
		return nil, err
	}
	// Nullable field processing
	if !botType.Valid {
		botType = sql.NullInt32{
			Int32: 0,
			Valid: true,
		}
	}
	if !botStateValue.Valid {
		botStateValue = sql.NullInt32{
			Int32: 1,
			Valid: true,
		}
	}

	return &pbDefiwallets.DeFiWallet{
		Recipient: &pbRecipients.Recipient{
			Id: getNullableString(botID),
		},
		Blockchain: &pbBlockchains.Blockchain{
			Id:      "",
			Name:    "",
			Type:    0,
			Icon:    nil,
			Evm:     false,
			Status:  0,
			Cryptos: nil,
		},
		Address: "",
	}, nil
}

func (s *DefiWalletRepository) QbUpdate(msg *pbDefiwallets.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _table)
	if msg.GetAddress() != "" {
		qb.SetUpdate("address", msg.GetAddress())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}
	switch msg.GetRecipient().GetSelect().GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where("core.defiwallets.id = ? ", msg.GetRecipient().GetSelect().GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		qb.Where(
			"core.defiwallets.id = (SELECT core.recipients.id FROM core.recipients "+
				"WHERE core.recipients.legalentity_id = ? "+
				"AND core.recipients.user_id = ? "+
				"AND core.recipients.label = ?)",
			msg.GetRecipient().GetSelect().GetByLegalEntityUserLabel().GetLegalEntity().GetById(),
			msg.GetRecipient().GetSelect().GetByLegalEntityUserLabel().GetUser().GetById(),
			msg.GetRecipient().GetSelect().GetByLegalEntityUserLabel().GetLabel(),
		)
	}

	return qb, nil
}

func (s *DefiWalletRepository) QbGetOne(_ *pbRecipients.GetRequest, recipientsQb *util.QueryBuilder) *util.QueryBuilder {
	recipientsQb.Join(fmt.Sprintf("LEFT JOIN %s ON defiwallets.id = recipients.id", _table))
	fields := strings.Split(_fields, ",")
	for i := 0; i < len(fields); i++ {
		fields[i] = _table + "." + fields[i]
	}
	newFields := strings.Join(fields, ",")
	recipientsQb.Select(newFields)
	return recipientsQb
}

func (s *DefiWalletRepository) QbGetList(msg *pbDefiwallets.GetListRequest, recipientsQb *util.QueryBuilder) *util.QueryBuilder {
	recipientsQb.Join(fmt.Sprintf("LEFT JOIN %s ON defiwallets.id = recipients.id", _table))
	fields := strings.Split(_fields, ",")
	for i := 0; i < len(fields); i++ {
		fields[i] = _table + "." + fields[i]
	}
	newFields := strings.Join(fields, ",")
	recipientsQb.Select(newFields)
	recipientsQb.Where("core.recipients.type = ?", pbRecipients.Type_TYPE_DV_BOT)
	if msg.Address != nil {
		recipientsQb.Where("core.defiwallets.address = ?", msg.GetAddress())
	}
	if msg.Blockchain != nil {
		recipientsQb.Where("core.defiwallets.blockchain_id = ?", msg.GetBlockchain())
	}

	return recipientsQb
}

func (s *DefiWalletRepository) ScanGetRow(row pgx.Row) (*pbDefiwallets.DeFiWallet, error) {
	var (
		recipientID           string
		nullableLegalentityID pgtype.Text
		nullableUserID        pgtype.Text
		recipientLabel        string
		recipientType         pbRecipients.Type
		nullableOrgID         pgtype.Text
		recipientStatus       pbCommon.Status
		botID                 string
		botType               sql.NullInt32
		defaultParamsName     sql.NullString
		botStateValue         sql.NullInt32
		_status               sql.NullInt32
	)

	err := row.Scan(
		&recipientID,
		&nullableLegalentityID,
		&nullableUserID,
		&recipientLabel,
		&recipientType,
		&nullableOrgID,
		&recipientStatus,
		&botID,
		&botType,
		&defaultParamsName,
		&botStateValue,
		&_status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing

	return &pbDefiwallets.DeFiWallet{
		Recipient: &pbRecipients.Recipient{
			Id:    recipientID,
			Label: recipientLabel,
			LegalEntity: &pbLegalEntities.LegalEntity{
				Id: nullableLegalentityID.String,
			},
			User: &pbUsers.User{
				Id: nullableUserID.String,
			},
			Org: &pbOrgs.Org{
				Id: nullableOrgID.String,
			},
			Type:   recipientType,
			Status: recipientStatus,
		},
		Blockchain: nil,
		Address:    "",
	}, nil
}

func (s *DefiWalletRepository) ScanListRow(row pgx.Row) (*pbDefiwallets.DeFiWallet, error) {
	var (
		recipientID           sql.NullString
		nullableLegalentityID sql.NullString
		nullableUserID        sql.NullString
		recipientLabel        sql.NullString
		recipientType         pbRecipients.Type
		nullableOrgID         sql.NullString
		recipientStatus       pbCommon.Status
		botID                 sql.NullString
		botType               sql.NullInt32
		defaultParamsName     sql.NullString
		botStateValue         sql.NullInt32
		_status               sql.NullInt32
	)

	err := row.Scan(
		&recipientID,
		&nullableLegalentityID,
		&nullableUserID,
		&recipientLabel,
		&recipientType,
		&nullableOrgID,
		&recipientStatus,
		&botID,
		&botType,
		&defaultParamsName,
		&botStateValue,
		&_status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing
	if !botType.Valid {
		botType = sql.NullInt32{
			Int32: 0,
			Valid: true,
		}
	}
	if !botStateValue.Valid {
		botStateValue = sql.NullInt32{
			Int32: 1,
			Valid: true,
		}
	}

	return &pbDefiwallets.DeFiWallet{
		Recipient: &pbRecipients.Recipient{
			Id:    getNullableString(recipientID),
			Label: getNullableString(recipientLabel),
			LegalEntity: &pbLegalEntities.LegalEntity{
				Id: getNullableString(nullableLegalentityID),
			},
			User: &pbUsers.User{
				Id: getNullableString(nullableUserID),
			},
			Org: &pbOrgs.Org{
				Id: getNullableString(nullableOrgID),
			},
			Type:   recipientType,
			Status: recipientStatus,
		},
		Blockchain: nil,
		Address:    "",
	}, nil
}

func (s *DefiWalletRepository) ScanWithRelationship(row pgx.Row) (*pbDataSources.DataSource, error) {
	var (
		dsID           string
		dsType         pbDataSources.Type
		dsName         string
		dsFsProviderID sql.NullString
		dsIcon         sql.NullString
		dsStatus       pbCommon.Status
	)

	var (
		fsproviderID     string
		fsproviderType   pbFsproviders.Type
		fsproviderName   string
		fsproviderIcon   sql.NullString
		fsproviderStatus pbCommon.Status
	)

	err := row.Scan(
		&dsID,
		&dsType,
		&dsName,
		&dsFsProviderID,
		&dsIcon,
		&dsStatus,
		&fsproviderID,
		&fsproviderType,
		&fsproviderName,
		&fsproviderIcon,
		&fsproviderStatus,
	)
	if err != nil {
		return nil, err
	}

	fmt.Println(fsproviderID, fsproviderType, fsproviderName)

	return &pbDataSources.DataSource{
		Id:   dsID,
		Type: dsType,
		Name: dsName,
		Provider: &pbFsproviders.FSProvider{
			Id:     fsproviderID,
			Type:   fsproviderType,
			Name:   fsproviderName,
			Icon:   util.GetSQLNullString(fsproviderIcon),
			Status: fsproviderStatus,
		},
		Icon:   util.GetSQLNullString(dsIcon),
		Status: dsStatus,
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
