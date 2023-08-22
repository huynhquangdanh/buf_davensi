package useraddresses

import (
	"errors"
	"reflect"
	"sync"

	pbCountries "davensi.com/core/gen/countries"

	pbAddresses "davensi.com/core/gen/addresses"
	pbKyc "davensi.com/core/gen/kyc"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName           = "core.users_addresses"
	_userAddressesFields = "user_id, label, address_id, main_address, ownership_status, status"
)

type UserAddressesRepository struct {
	db *pgxpool.Pool
}

var (
	singleRepo *UserAddressesRepository
	once       sync.Once
)

func NewUserAddressRepository(db *pgxpool.Pool) *UserAddressesRepository {
	return &UserAddressesRepository{
		db: db,
	}
}

func GetSingletonRepository(db *pgxpool.Pool) *UserAddressesRepository {
	once.Do(func() {
		singleRepo = NewUserAddressRepository(db)
	})
	return singleRepo
}

func (*UserAddressesRepository) QbUpsertUserAddresses(
	userID string,
	addresses []*pbAddresses.SetLabeledAddress,
	command string,
) (*util.QueryBuilder, error) {
	var queryType util.QueryType
	if command == "upsert" {
		queryType = util.Upsert
	} else {
		queryType = util.Insert
	}
	qb := util.CreateQueryBuilder(queryType, _tableName)
	qb.SetInsertField("user_id", "label", "address_id", "main_address", "ownership_status", "status")

	for _, address := range addresses {
		_, err := qb.SetInsertValues([]any{
			userID,
			address.GetLabel(),
			address.GetId(),
			address.GetMainAddress(),
			address.GetOwnershipStatus().Enum(),
			address.GetStatus().Enum(),
		})
		if err != nil {
			return nil, err
		}
	}

	return qb, nil
}

func (*UserAddressesRepository) QbGetUserAddressByIDLabel(userID, label string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)

	qb.Select(_userAddressesFields)
	qb.Where("user_id = ? AND label = ? AND status != ?", userID, label, pbKyc.Status_STATUS_CANCELED)

	return qb
}

func (r *UserAddressesRepository) ScanRow(row pgx.Rows) ([]*pbAddresses.LabeledAddress, error) {
	var result []*pbAddresses.LabeledAddress
	for row.Next() {
		var (
			userID          string
			label           string
			addressID       string
			mainAddress     bool
			ownerShipStatus *pbAddresses.OwnershipStatus
			status          *pbKyc.Status
		)
		err := row.Scan(&userID, &label, &addressID, &mainAddress, &ownerShipStatus, &status)
		if err != nil {
			return nil, err
		}
		result = append(result, &pbAddresses.LabeledAddress{
			Label: label,
			Address: &pbAddresses.Address{
				Id: addressID,
			},
			MainAddress:     &mainAddress,
			OwnershipStatus: ownerShipStatus,
			Status:          *status.Enum(),
		})
	}
	return result, nil
}

func (*UserAddressesRepository) QbUpdate(
	userID string, updateRequest *pbAddresses.UpdateLabeledAddressRequest,
) (qb *util.QueryBuilder, err error) {
	// TODO: all fields are relationship fields, do later

	qb = util.CreateQueryBuilder(util.Update, _tableName)
	if updateRequest.GetStatus() != pbKyc.Status_STATUS_UNSPECIFIED {
		qb.SetUpdate("status", updateRequest.GetStatus())
	}
	if !reflect.ValueOf(updateRequest.MainAddress).IsNil() {
		qb.SetUpdate("main_address", updateRequest.GetMainAddress())
	}
	if !reflect.ValueOf(updateRequest.OwnershipStatus).IsNil() {
		qb.SetUpdate("ownership_status", updateRequest.GetOwnershipStatus())
	}
	if !reflect.ValueOf(updateRequest.Status).IsNil() {
		qb.SetUpdate("status", updateRequest.GetStatus())
	}
	if !reflect.ValueOf(updateRequest.Address).IsNil() && !reflect.ValueOf(updateRequest.Address.Label).IsNil() {
		qb.SetUpdate("label", updateRequest.GetAddress().GetLabel())
	}
	qb.Where("label = ?", updateRequest.GetByLabel())

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("user_id = ?", userID)

	return qb, nil
}

func (*UserAddressesRepository) CheckUpdateability(
	oldUserAddress *pbAddresses.LabeledAddress,
	updateUserAddressRequest *pbAddresses.UpdateLabeledAddressRequest,
) (isUpdateable, isSubAddressUpdateable bool) {
	isUpdateable = true

	if !reflect.ValueOf(updateUserAddressRequest.MainAddress).IsNil() {
		isUpdateable = oldUserAddress.GetMainAddress() != updateUserAddressRequest.GetMainAddress()
	}
	if !reflect.ValueOf(updateUserAddressRequest.Status).IsNil() {
		isUpdateable = oldUserAddress.GetStatus() != updateUserAddressRequest.GetStatus()
	}
	if !reflect.ValueOf(updateUserAddressRequest.OwnershipStatus).IsNil() {
		isUpdateable = oldUserAddress.GetOwnershipStatus() != updateUserAddressRequest.GetOwnershipStatus()
	}

	if reflect.ValueOf(updateUserAddressRequest.Address).IsNil() {
		return isUpdateable, isSubAddressUpdateable
	}
	compareAndUpdate := func(oldFieldValue, newFieldValue, newField any) {
		if !reflect.ValueOf(newField).IsNil() && !reflect.DeepEqual(oldFieldValue, newFieldValue) {
			isUpdateable = !reflect.DeepEqual(oldFieldValue, newFieldValue)
			isSubAddressUpdateable = true
		}
	}
	updateAddress := updateUserAddressRequest.GetAddress()
	oldUserAddressDetail := oldUserAddress.GetAddress()
	if !reflect.ValueOf(updateAddress.Label).IsNil() {
		isUpdateable = oldUserAddress.GetLabel() != updateAddress.GetLabel()
	}
	if !reflect.ValueOf(updateAddress.Country).IsNil() {
		switch updateAddress.GetCountry().GetSelect().(type) {
		case *pbCountries.Select_ById:
			isUpdateable = oldUserAddress.GetAddress().GetCountry().GetId() != updateAddress.GetCountry().GetById()
		case *pbCountries.Select_ByCode:
			isUpdateable = oldUserAddress.GetAddress().GetCountry().GetCode() != updateAddress.GetCountry().GetByCode()
		}
		isSubAddressUpdateable = true
	}
	compareAndUpdate(oldUserAddressDetail.GetType(), updateAddress.GetType(), updateAddress.Type)
	compareAndUpdate(oldUserAddressDetail.GetBuilding(), updateAddress.GetBuilding(), updateAddress.Building)
	compareAndUpdate(oldUserAddressDetail.GetFloor(), updateAddress.GetFloor(), updateAddress.Floor)
	compareAndUpdate(oldUserAddressDetail.GetUnit(), updateAddress.GetUnit(), updateAddress.Unit)
	compareAndUpdate(oldUserAddressDetail.GetStreetNum(), updateAddress.GetStreetNum(), updateAddress.StreetNum)
	compareAndUpdate(oldUserAddressDetail.GetStreetName(), updateAddress.GetStreetName(), updateAddress.StreetName)
	compareAndUpdate(oldUserAddressDetail.GetDistrict(), updateAddress.GetDistrict(), updateAddress.District)
	compareAndUpdate(oldUserAddressDetail.GetLocality(), updateAddress.GetLocality(), updateAddress.Locality)
	compareAndUpdate(oldUserAddressDetail.GetZipCode(), updateAddress.GetZipCode(), updateAddress.ZipCode)
	compareAndUpdate(oldUserAddressDetail.GetRegion(), updateAddress.GetRegion(), updateAddress.Region)
	compareAndUpdate(oldUserAddressDetail.GetState(), updateAddress.GetState(), updateAddress.State)
	compareAndUpdate(oldUserAddressDetail.GetStatus(), updateAddress.GetStatus(), updateAddress.Status)

	return isUpdateable, isSubAddressUpdateable
}

func (*UserAddressesRepository) ScanMultiRows(rows pgx.Rows) (*pbAddresses.LabeledAddressList, error) {
	var (
		userID          string
		addressID       pgtype.Text
		label           string
		mainAddress     bool
		ownershipStatus int32
		status          uint32
	)

	userAddresss := []*pbAddresses.LabeledAddress{}

	for rows.Next() {
		err := rows.Scan(
			&userID,
			&label,
			&addressID,
			&mainAddress,
			&ownershipStatus,
			&status,
		)
		if err != nil {
			return nil, err
		}
		userAddresss = append(userAddresss, &pbAddresses.LabeledAddress{
			Label: label,
			Address: &pbAddresses.Address{
				Id: addressID.String,
			},
			MainAddress:     &mainAddress,
			OwnershipStatus: pbAddresses.OwnershipStatus(ownershipStatus).Enum(),
			Status:          pbKyc.Status(status),
		})
	}
	return &pbAddresses.LabeledAddressList{
		List: userAddresss,
	}, nil
}
