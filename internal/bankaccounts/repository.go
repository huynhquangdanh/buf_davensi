package bankaccounts

import (
	"errors"
	"fmt"
	"strings"

	pbAddresses "davensi.com/core/gen/addresses"
	pbBankAccounts "davensi.com/core/gen/bankaccounts"
	pbBankAccountsConnect "davensi.com/core/gen/bankaccounts/bankaccountsconnect"
	pbBankBranches "davensi.com/core/gen/bankbranches"
	pbBanks "davensi.com/core/gen/banks"
	"davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbOrgs "davensi.com/core/gen/orgs"
	pbRecipients "davensi.com/core/gen/recipients"
	pbUoms "davensi.com/core/gen/uoms"
	pbUsers "davensi.com/core/gen/users"

	"davensi.com/core/internal/legalentities"
	"davensi.com/core/internal/users"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.bankaccounts"
	_fields    = "id, bankbranch_id, bankaccount_type, currency_id, pan, masked_pan, bban, iban, external_id"
)

type BankAccountRepository struct {
	pbBankAccountsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewBankRepository(db *pgxpool.Pool) *BankAccountRepository {
	return &BankAccountRepository{
		db: db,
	}
}

func (s *BankAccountRepository) QbInsert(msg *pbBankAccounts.CreateRequest, recipientID string) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleBankAccountValue := []any{}

	qb.SetInsertField("id")
	singleBankAccountValue = append(singleBankAccountValue, recipientID)

	// Append optional fields values
	if msg.BankBranch != nil && msg.GetBankBranch().GetById() != "" {
		qb.SetInsertField("bankbranch_id")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetBankBranch().GetById())
	} else if msg.BankBranch == nil || msg.GetBankBranch().GetById() == "" {
		qb.SetInsertField("bankbranch_id")
		singleBankAccountValue = append(singleBankAccountValue, "00000000-0000-0000-0000-000000000000")
	}

	if msg.BankAccountType != nil {
		qb.SetInsertField("bankaccount_type")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetBankAccountType())
	}

	if msg.Currency != nil && msg.GetCurrency().GetById() != "" {
		qb.SetInsertField("currency_id")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetCurrency().GetById())
	}

	if msg.Pan != nil {
		qb.SetInsertField("pan")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetPan())
	}

	if msg.MaskedPan != nil {
		qb.SetInsertField("masked_pan")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetMaskedPan())
	}

	if msg.Bban != nil {
		qb.SetInsertField("bban")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetBban())
	}

	if msg.Iban != nil {
		qb.SetInsertField("iban")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetIban())
	}

	if msg.ExternalId != nil {
		qb.SetInsertField("external_id")
		singleBankAccountValue = append(singleBankAccountValue, msg.GetExternalId())
	}

	_, err := qb.SetInsertValues(singleBankAccountValue)

	return qb, err
}

func (s *BankAccountRepository) QbGetOne(req *pbRecipients.GetRequest) *util.QueryBuilder {
	qb := util.
		CreateQueryBuilder(util.Select, _tableName).
		Select(util.GetFieldsWithTableName(_fields, _tableName))

	SetQBBySelect(req.GetSelect(), qb)

	return qb
}

func SetQBBySelect(selectRecipient *pbRecipients.Select, qb *util.QueryBuilder) {
	switch selectRecipient.GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where("bankaccounts.id = ?", selectRecipient.GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		filterArgs := []any{}

		qbGetLegalEntity := util.CreateQueryBuilder(util.Select, "core.legalentities")
		legalentities.SetQBBySelect(
			selectRecipient.GetByLegalEntityUserLabel().GetLegalEntity(),
			qbGetLegalEntity.Select("id"),
		)
		legalEntitySQL, legalEntityArgs, _ := qbGetLegalEntity.GenerateSQL()

		qbGetUser := util.CreateQueryBuilder(util.Select, "core.users")
		users.SetQBBySelect(
			selectRecipient.GetByLegalEntityUserLabel().GetUser(),
			qbGetUser.Select("id"),
		)
		userSQL, userArgs, _ := qbGetUser.GenerateSQL()

		filterArgs = append(filterArgs, legalEntityArgs...)
		filterArgs = append(filterArgs, userArgs...)
		filterArgs = append(filterArgs, selectRecipient.GetByLegalEntityUserLabel().GetLabel())

		if selectRecipient.GetByLegalEntityUserLabel().GetLegalEntity() == nil {
			legalEntitySQL = "SELECT id FROM core.legalentities"
		}

		if selectRecipient.GetByLegalEntityUserLabel().GetUser() == nil {
			userSQL = "SELECT id FROM core.users"
		}

		qb.Where(
			fmt.Sprintf(
				"bankaccounts.id = (SELECT id FROM core.recipients WHERE "+
					"recipients.legalentity_id IN (%s) AND recipients.user_id IN (%s) AND recipients.label = ?)", legalEntitySQL, userSQL,
			),
			filterArgs...)
	}
}

func (s *BankAccountRepository) QbUpdate(req *pbBankAccounts.UpdateRequest, bankAccountID string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if req.BankBranch != nil {
		qb.SetUpdate("bankbranch_id", req.GetBankBranch().GetById())
	}

	if req.BankAccountType != nil {
		qb.SetUpdate("bankaccount_type", req.GetBankAccountType())
	}

	if req.Currency != nil {
		qb.SetUpdate("currency_id", req.GetCurrency().GetById())
	}

	if req.Pan != nil {
		qb.SetUpdate("pan", req.GetPan())
	}

	if req.MaskedPan != nil {
		qb.SetUpdate("masked_pan", req.GetMaskedPan())
	}

	if req.Bban != nil {
		qb.SetUpdate("bban", req.GetBban())
	}

	if req.Iban != nil {
		qb.SetUpdate("iban", req.GetIban())
	}

	if req.ExternalId != nil {
		qb.SetUpdate("external_id", req.GetExternalId())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", bankAccountID)

	return qb, nil
}

func (s *BankAccountRepository) QbGetList(req *pbBankAccounts.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, "bankaccounts"))

	if req.BankAccountType != nil {
		bankAccountTypes := req.GetBankAccountType().GetList()

		if len(bankAccountTypes) > 0 {
			args := []any{}
			for _, v := range bankAccountTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"bankaccount_type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(bankAccountTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if req.Pan != nil {
		qb.Where("pan LIKE '%' || ? || '%'", req.GetPan())
	}

	if req.MaskedPan != nil {
		qb.Where("masked_pan LIKE '%' || ? || '%'", req.GetMaskedPan())
	}

	if req.Bban != nil {
		qb.Where("bban LIKE '%' || ? || '%'", req.GetBban())
	}

	if req.Iban != nil {
		qb.Where("iban LIKE '%' || ? || '%'", req.GetIban())
	}

	if req.ExternalId != nil {
		qb.Where("external_id LIKE '%' || ? || '%'", req.GetExternalId())
	}

	return qb
}

func ScanRow(row pgx.Row) (*pbBankAccounts.BankAccount, error) {
	var ( // main table fields
		id              string
		bankBranchID    string
		bankAccountType uint32
		currencyID      pgtype.Text
		pan             pgtype.Text
		maskedPan       pgtype.Text
		bban            pgtype.Text
		iban            pgtype.Text
		externalID      pgtype.Text
	)

	err := row.Scan(
		&id,
		&bankBranchID,
		&bankAccountType,
		&currencyID,
		&pan,
		&maskedPan,
		&bban,
		&iban,
		&externalID,
	)
	if err != nil {
		return nil, err
	}

	return &pbBankAccounts.BankAccount{
		Recipient: &pbRecipients.Recipient{
			Id: id,
		},
		BankBranch: &pbBankBranches.BankBranch{
			Id: bankBranchID,
		},
		BankAccountType: pbBankAccounts.Type(bankAccountType),
		Currency: &pbUoms.UoM{
			Id: currencyID.String,
		},
		Pan:        pan.String,
		MaskedPan:  &maskedPan.String,
		Bban:       &bban.String,
		Iban:       &iban.String,
		ExternalId: &externalID.String,
	}, nil
}

func (s *BankAccountRepository) ScanWithRelationship(row pgx.Row) (*pbBankAccounts.BankAccount, error) {
	var ( // main table fields
		id              string
		bankBranchID    string
		bankAccountType uint32
		currencyID      pgtype.Text
		pan             pgtype.Text
		maskedPan       pgtype.Text
		bban            pgtype.Text
		iban            pgtype.Text
		externalID      pgtype.Text
	)

	var ( // recipients fields
		recipientID            string
		recipientLegalEntityID pgtype.Text
		recipientUserID        pgtype.Text
		recipientLabel         pgtype.Text
		recipientType          pgtype.Int2
		recipientOrgID         pgtype.Text
		recipientStatus        pgtype.Int2
	)

	var (
		bankBranchBankBranchID pgtype.Text
		bankBranchBankID       pgtype.Text
		bankBranchBranchCode   pgtype.Text
		bankBranchType         pgtype.Int2
		bankBranchName         pgtype.Text
		bankBranchAddressID    pgtype.Text
		bankBranchContact1ID   pgtype.Text
		bankBranchContact2ID   pgtype.Text
		bankBranchContact3ID   pgtype.Text
		bankBranchStatus       pgtype.Int2
	)

	var ( // uoms fields
		uomID                 pgtype.Text
		uomType               pgtype.Int2
		uomSymbol             pgtype.Text
		uomName               pgtype.Text
		uomIcon               pgtype.Text
		uomManagedDecimals    pgtype.Int2
		uomDisplayedDecipmals pgtype.Int2
		uomReportingUnit      pgtype.Bool
		uomStatus             pgtype.Int2
	)

	err := row.Scan(
		&id,
		&bankBranchID,
		&bankAccountType,
		&currencyID,
		&pan,
		&maskedPan,
		&bban,
		&iban,
		&externalID,
		&recipientID,
		&recipientLegalEntityID,
		&recipientUserID,
		&recipientLabel,
		&recipientType,
		&recipientOrgID,
		&recipientStatus,
		&bankBranchBankBranchID,
		&bankBranchBankID,
		&bankBranchBranchCode,
		&bankBranchType,
		&bankBranchName,
		&bankBranchAddressID,
		&bankBranchContact1ID,
		&bankBranchContact2ID,
		&bankBranchContact3ID,
		&bankBranchStatus,
		&uomID,
		&uomType,
		&uomSymbol,
		&uomName,
		&uomIcon,
		&uomManagedDecimals,
		&uomDisplayedDecipmals,
		&uomReportingUnit,
		&uomStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbBankAccounts.BankAccount{
		Recipient: &pbRecipients.Recipient{
			Id: recipientID,
			LegalEntity: &pbLegalEntities.LegalEntity{
				Id: recipientLegalEntityID.String,
			},
			User: &pbUsers.User{
				Id: recipientUserID.String,
			},
			Label: recipientLabel.String,
			Type:  pbRecipients.Type(recipientType.Int16),
			Org: &pbOrgs.Org{
				Id: recipientOrgID.String,
			},
			Status: common.Status(recipientStatus.Int16),
		},
		BankBranch: &pbBankBranches.BankBranch{
			Id: bankBranchBankBranchID.String,
			Bank: &pbBanks.Bank{
				Id: bankBranchBankID.String,
			},
			BranchCode: bankBranchBranchCode.String,
			Type:       pbBanks.Type(bankBranchType.Int16),
			Name:       bankBranchName.String,
			Address: &pbAddresses.Address{
				Id: bankBranchAddressID.String,
			},
			Contact1: &pbContacts.Contact{
				Id: bankBranchContact1ID.String,
			},
			Contact2: &pbContacts.Contact{
				Id: bankBranchContact2ID.String,
			},
			Contact3: &pbContacts.Contact{
				Id: bankBranchContact3ID.String,
			},
			Status: common.Status(bankBranchStatus.Int16),
		},
		BankAccountType: pbBankAccounts.Type(bankAccountType),
		Currency: &pbUoms.UoM{
			Id:                uomID.String,
			Type:              pbUoms.Type(uomType.Int16),
			Symbol:            uomSymbol.String,
			Name:              &uomName.String,
			Icon:              &uomIcon.String,
			ManagedDecimals:   uint32(uomManagedDecimals.Int16),
			DisplayedDecimals: uint32(uomDisplayedDecipmals.Int16),
			ReportingUnit:     uomReportingUnit.Bool,
			Status:            common.Status(uomStatus.Int16),
		},
		Pan:        pan.String,
		MaskedPan:  &maskedPan.String,
		Bban:       &bban.String,
		Iban:       &iban.String,
		ExternalId: &externalID.String,
	}, nil
}
