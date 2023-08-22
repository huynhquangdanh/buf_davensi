package bankbranches

import (
	"errors"
	"fmt"
	"strings"

	pbAddresses "davensi.com/core/gen/addresses"
	pbBankBranches "davensi.com/core/gen/bankbranches"
	pbBankBranchesConnect "davensi.com/core/gen/bankbranches/bankbranchesconnect"
	pbBanks "davensi.com/core/gen/banks"
	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbCountries "davensi.com/core/gen/countries"

	"davensi.com/core/internal/banks"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName          = "core.bankbranches"
	_bankbranchesFields = "id, bank_id, branch_code, type, name, address_id, contact1_id, contact2_id, contact3_id, status"
)

type BankBranchRepository struct {
	pbBankBranchesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewBankRepository(db *pgxpool.Pool) *BankBranchRepository {
	return &BankBranchRepository{
		db: db,
	}
}

func (s *BankBranchRepository) QbInsert(req *pbBankBranches.CreateRequest, addressID, contact1ID, contact2ID, contact3ID string) (
	*util.QueryBuilder, error,
) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleBankBranchValue := []any{}

	// Append required fields
	qb.SetInsertField("bank_id")
	singleBankBranchValue = append(singleBankBranchValue, req.GetBank().GetById())
	qb.SetInsertField("branch_code")
	singleBankBranchValue = append(singleBankBranchValue, req.GetBranchCode())
	qb.SetInsertField("type")
	singleBankBranchValue = append(singleBankBranchValue, req.GetType())
	qb.SetInsertField("name")
	singleBankBranchValue = append(singleBankBranchValue, req.GetName())

	// Append optional fields values
	if req.Status != nil {
		qb.SetInsertField("status")
		singleBankBranchValue = append(singleBankBranchValue, req.GetStatus())
	}

	if req.Address != nil {
		qb.SetInsertField("address_id")
		singleBankBranchValue = append(singleBankBranchValue, addressID)
	}

	if req.Contact1 != nil {
		qb.SetInsertField("contact1_id")
		singleBankBranchValue = append(singleBankBranchValue, contact1ID)
	}

	if req.Contact2 != nil {
		qb.SetInsertField("contact2_id")
		singleBankBranchValue = append(singleBankBranchValue, contact2ID)
	}

	if req.Contact3 != nil {
		qb.SetInsertField("contact3_id")
		singleBankBranchValue = append(singleBankBranchValue, contact3ID)
	}

	_, err := qb.SetInsertValues(singleBankBranchValue)

	return qb, err
}

func (s *BankBranchRepository) QbUpdate(req *pbBankBranches.UpdateRequest,
	addressUUID, contact1UUID, contact2UUID, contact3UUID string,
) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	// TODO: update bank_code, bic, contact1,2,3
	if req.Name != nil {
		qb.SetUpdate("name", req.GetName())
	}

	if req.Type != nil {
		qb.SetUpdate("type", req.GetType())
	}

	if req.Status != nil {
		qb.SetUpdate("status", req.GetStatus())
	}

	// Update address_ID: update to same existing ID if already linked, else write new
	if req.Address != nil {
		qb.SetUpdate("address_id", addressUUID)
	}

	// Update contact1_ID: update to same existing ID if already linked, else write new
	if req.Contact1 != nil {
		qb.SetUpdate("contact1_id", contact1UUID)
	}

	if req.Contact2 != nil {
		qb.SetUpdate("contact2_id", contact2UUID)
	}

	if req.Contact3 != nil {
		qb.SetUpdate("contact3_id", contact3UUID)
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	SetQBBySelect(req.GetSelect(), qb)

	return qb, nil
}

func (s *BankBranchRepository) QbGetOne(req *pbBankBranches.GetRequest) *util.QueryBuilder {
	qb := util.
		CreateQueryBuilder(util.Select, _tableName).
		Select(util.GetFieldsWithTableName(_bankbranchesFields, _tableName)).
		Where("bankbranches.status = ?", pbCommon.Status_STATUS_ACTIVE)

	SetQBBySelect(req.GetSelect(), qb)

	return qb
}

func SetQBBySelect(selectBankBranch *pbBankBranches.Select, qb *util.QueryBuilder) {
	switch selectBankBranch.GetSelect().(type) {
	case *pbBankBranches.Select_ById:
		qb.Where("bankbranches.id = ?", selectBankBranch.GetById())
	case *pbBankBranches.Select_ByBankBranchCode:
		filterArgs := []any{}

		qbGetBank := util.CreateQueryBuilder(util.Select, "core.banks")
		banks.SetQBBySelect(
			selectBankBranch.GetByBankBranchCode().GetBank(),
			qbGetBank.Select("id"),
		)
		bankSQL, bankArgs, _ := qbGetBank.GenerateSQL()

		filterArgs = append(filterArgs, bankArgs...)
		filterArgs = append(filterArgs, selectBankBranch.GetByBankBranchCode().GetBranchCode())

		qb.Where(fmt.Sprintf("bankbranches.bank_id IN (%s) AND bankbranches.branch_code =?", bankSQL), filterArgs...)
	}
}

func (s *BankBranchRepository) QbGetList(req *pbBankBranches.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_bankbranchesFields, "bankbranches"))

	if req.BranchCode != nil {
		qb.Where("bankbranches.branch_code LIKE '%' || ? || '%'", req.GetBranchCode())
	}

	if req.Type != nil {
		bankBranchTypes := req.GetType().GetList()

		if len(bankBranchTypes) > 0 {
			args := []any{}
			for _, v := range bankBranchTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"bankbranches.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(bankBranchTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if req.Name != nil {
		qb.Where("bankbranches.name LIKE '%' || ? || '%'", req.GetName())
	}

	if req.Status != nil {
		bankBranchStatuses := req.GetStatus().GetList()

		if len(bankBranchStatuses) > 0 {
			args := []any{}
			for _, v := range bankBranchStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"bankbranches.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(bankBranchStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func ScanRow(row pgx.Row) (*pbBankBranches.BankBranch, error) {
	var (
		id             pgtype.Text
		bankID         pgtype.Text
		branchCode     pgtype.Text
		bankBranchType uint32
		name           pgtype.Text
		addressID      pgtype.Text
		contact1ID     pgtype.Text
		contact2ID     pgtype.Text
		contact3ID     pgtype.Text
		status         uint32
	)

	err := row.Scan(
		&id,
		&bankID,
		&branchCode,
		&bankBranchType,
		&name,
		&addressID,
		&contact1ID,
		&contact2ID,
		&contact3ID,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbBankBranches.BankBranch{
		Id: id.String,
		Bank: &pbBanks.Bank{
			Id: bankID.String,
		},
		BranchCode: branchCode.String,
		Type:       pbBanks.Type(bankBranchType),
		Name:       name.String,
		Address: &pbAddresses.Address{
			Id: addressID.String,
		},
		Contact1: &pbContacts.Contact{
			Id: contact1ID.String,
		},
		Contact2: &pbContacts.Contact{
			Id: contact2ID.String,
		},
		Contact3: &pbContacts.Contact{
			Id: contact3ID.String,
		},
		Status: pbCommon.Status(status),
	}, nil
}

func (s *BankBranchRepository) ScanWithRelationship(row pgx.Row) (*pbBankBranches.BankBranch, error) {
	var ( // main table fields
		id             string
		bankID         string
		branchCode     string
		bankBranchType uint32
		name           string
		addressID      pgtype.Text
		contact1ID     pgtype.Text
		contact2ID     pgtype.Text
		contact3ID     pgtype.Text
		status         uint32
	)

	var ( // bank fields
		bankBankID             pgtype.Text
		bankName               pgtype.Text
		bankType               pgtype.Int2
		bankBic                pgtype.Text
		bankBankCode           pgtype.Text
		bankAdressID           pgtype.Text
		bankContact1ID         pgtype.Text
		bankContact2ID         pgtype.Text
		bankContact3ID         pgtype.Text
		bankOpenBankingSupport pgtype.Bool
		bankParentID           pgtype.Text
		bankStatus             pgtype.Int2
	)

	var ( // address fields
		addressAddressID  pgtype.Text
		addressType       pgtype.Int2
		addressCountryID  pgtype.Text
		addressBuilding   pgtype.Text
		addressFloor      pgtype.Text
		addressUnit       pgtype.Text
		addressStreetNum  pgtype.Text
		addressStreetName pgtype.Text
		addressDistrict   pgtype.Text
		addressLocality   pgtype.Text
		addressZipCode    pgtype.Text
		addressRegion     pgtype.Text
		addressState      pgtype.Text
		addressStatus     pgtype.Int2
	)

	var ( // contact1 fields
		contactContact1ID pgtype.Text
		contact1Type      pgtype.Int2
		contact1Value     pgtype.Text
		contact1Status    pgtype.Int2
	)

	var ( // contact2 fields
		contactContact2ID pgtype.Text
		contact2Type      pgtype.Int2
		contact2Value     pgtype.Text
		contact2Status    pgtype.Int2
	)

	var ( // contact3 fields
		contactContact3ID pgtype.Text
		contact3Type      pgtype.Int2
		contact3Value     pgtype.Text
		contact3Status    pgtype.Int2
	)

	err := row.Scan(
		&id,
		&bankID,
		&branchCode,
		&bankBranchType,
		&name,
		&addressID,
		&contact1ID,
		&contact2ID,
		&contact3ID,
		&status,
		&bankBankID,
		&bankName,
		&bankType,
		&bankBic,
		&bankBankCode,
		&bankAdressID,
		&bankContact1ID,
		&bankContact2ID,
		&bankContact3ID,
		&bankOpenBankingSupport,
		&bankParentID,
		&bankStatus,
		&addressAddressID,
		&addressType,
		&addressCountryID,
		&addressBuilding,
		&addressFloor,
		&addressUnit,
		&addressStreetNum,
		&addressStreetName,
		&addressDistrict,
		&addressLocality,
		&addressZipCode,
		&addressRegion,
		&addressState,
		&addressStatus,
		&contactContact1ID,
		&contact1Type,
		&contact1Value,
		&contact1Status,
		&contactContact2ID,
		&contact2Type,
		&contact2Value,
		&contact2Status,
		&contactContact3ID,
		&contact3Type,
		&contact3Value,
		&contact3Status,
	)
	if err != nil {
		return nil, err
	}

	return &pbBankBranches.BankBranch{
		Id: id,
		Bank: &pbBanks.Bank{
			Id:       bankBankID.String,
			Name:     bankName.String,
			Type:     pbBanks.Type(bankType.Int16),
			Bic:      bankBic.String,
			BankCode: bankBankCode.String,
			Address: &pbAddresses.Address{
				Id: bankAdressID.String,
			},
			Contact1: &pbContacts.Contact{
				Id: bankContact1ID.String,
			},
			Contact2: &pbContacts.Contact{
				Id: bankContact2ID.String,
			},
			Contact3: &pbContacts.Contact{
				Id: bankContact3ID.String,
			},
			OpenbankingSupport: &bankOpenBankingSupport.Bool,
			Parent: &pbBanks.Bank{
				Id: bankParentID.String,
			},
			Status: pbCommon.Status(bankStatus.Int16),
		},
		BranchCode: branchCode,
		Type:       pbBanks.Type(bankBranchType),
		Name:       name,
		Address: &pbAddresses.Address{
			Id:   addressAddressID.String,
			Type: pbAddresses.Type(addressType.Int16),
			Country: &pbCountries.Country{
				Id: addressCountryID.String,
			},
			Building:   &addressBuilding.String,
			Floor:      &addressFloor.String,
			Unit:       &addressUnit.String,
			StreetNum:  &addressStreetNum.String,
			StreetName: &addressStreetName.String,
			District:   &addressDistrict.String,
			Locality:   &addressLocality.String,
			ZipCode:    &addressZipCode.String,
			Region:     &addressRegion.String,
			State:      &addressState.String,
			Status:     pbCommon.Status(addressStatus.Int16),
		},
		Contact1: &pbContacts.Contact{
			Id:     contactContact1ID.String,
			Type:   pbContacts.Type(contact1Type.Int16),
			Value:  contact1Value.String,
			Status: pbCommon.Status(contact1Status.Int16),
		},
		Contact2: &pbContacts.Contact{
			Id:     contactContact2ID.String,
			Type:   pbContacts.Type(contact2Type.Int16),
			Value:  contact2Value.String,
			Status: pbCommon.Status(contact2Status.Int16),
		},
		Contact3: &pbContacts.Contact{
			Id:     contactContact3ID.String,
			Type:   pbContacts.Type(contact3Type.Int16),
			Value:  contact3Value.String,
			Status: pbCommon.Status(contact3Status.Int16),
		},
		Status: pbCommon.Status(status),
	}, nil
}

// The query for GETLIST() use multiple JOINs on one same table, thus the need to modify the query after generating it
// This is for GETLIST() only because it allows to find records with status != 1
func generateAndTransformSQL(qb *util.QueryBuilder) (sqlStr string, sqlArgs []any) {
	replaceCount := 4 // number of times EACH 'contacts.{column}' apppears, to replace with aliases later

	sqlStr, sqlArgs, _ = qb.GenerateSQL()
	// Transform sqlstr to give aliases to core.uoms fields because there are multiple left-joins on that table
	sqlStr = strings.Replace(sqlStr, "contacts", "contact1", replaceCount)
	sqlStr = strings.Replace(sqlStr, "contacts", "contact2", replaceCount)
	sqlStr = strings.Replace(sqlStr, "contacts", "contact3", replaceCount)
	sqlStr = strings.Replace(sqlStr, "core.contacts ON bankbranches.contact1_id = contacts.id",
		"core.contacts contact1 ON bankbranches.contact1_id = contact1.id", 1)
	sqlStr = strings.Replace(sqlStr, "core.contacts ON bankbranches.contact2_id = contacts.id",
		"core.contacts contact2 ON bankbranches.contact2_id = contact2.id", 1)
	sqlStr = strings.Replace(sqlStr, "core.contacts ON bankbranches.contact3_id = contacts.id",
		"core.contacts contact3 ON bankbranches.contact3_id = contact3.id", 1)

	sqlStr = strings.Replace(sqlStr, "SELECT ", "SELECT DISTINCT ", 1)

	return sqlStr, sqlArgs
}
