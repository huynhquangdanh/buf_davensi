package legalentities

import (
	"context"

	pbAddresses "davensi.com/core/gen/addresses"
	pbContacts "davensi.com/core/gen/contacts"
	pbCountries "davensi.com/core/gen/countries"
	pbKYC "davensi.com/core/gen/kyc"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbUoms "davensi.com/core/gen/uoms"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type LegalEntityRelationships struct {
	Country *pbCountries.Country
	UoM1    *pbUoms.UoM
	UoM2    *pbUoms.UoM
	UoM3    *pbUoms.UoM
}

// Backed by table core.legalentities_addresses. This is for getting address list for Legal Entity only
type AddressInfo struct {
	Label           string
	AddressID       string
	MainAddress     bool
	OwnershipStatus int16
	Status          int16
}

// Backed by table core.legalentities_contacts. This is for getting contact list for Legal Entity only
type ContactInfo struct {
	Label       string
	ContactID   string
	MainContact bool
	Status      int16
}

// For Incorporation Country, Currency1, Currency2, Currency3
func (s *ServiceServer) GetRelationship(
	selectCountry *pbCountries.Select,
	selectUom1,
	selectUom2,
	selectUom3 *pbUoms.Select,
) LegalEntityRelationships {
	countryChan := make(chan *pbCountries.Country)
	UomChan1, UomChan2, UomChan3 := make(chan *pbUoms.UoM), make(chan *pbUoms.UoM), make(chan *pbUoms.UoM)

	// Incorporation country
	go func() {
		var existCountry *pbCountries.Country
		if selectCountry != nil {
			getCountryResponse, err := s.countriesSS.Get(context.Background(), &connect.Request[pbCountries.GetRequest]{
				Msg: &pbCountries.GetRequest{
					Select: selectCountry,
				},
			})
			if err == nil {
				existCountry = getCountryResponse.Msg.GetCountry()
			}
		}
		countryChan <- existCountry
	}()

	// Currency1 (required)
	go func() {
		var existUom *pbUoms.UoM
		if selectUom1 != nil {
			getUomResponse, err := s.uomsSS.Get(context.Background(), &connect.Request[pbUoms.GetRequest]{
				Msg: &pbUoms.GetRequest{
					Select: selectUom1,
				},
			})
			if err == nil {
				existUom = getUomResponse.Msg.GetUom()
			}
		}
		UomChan1 <- existUom
	}()

	// Currency2 (optional)
	go func() {
		var existUom *pbUoms.UoM
		if selectUom2 != nil {
			getUomResponse, err := s.uomsSS.Get(context.Background(), &connect.Request[pbUoms.GetRequest]{
				Msg: &pbUoms.GetRequest{
					Select: selectUom2,
				},
			})
			if err == nil {
				existUom = getUomResponse.Msg.GetUom()
			}
		}
		UomChan2 <- existUom
	}()

	// Currency3 (optional)
	go func() {
		var existUom *pbUoms.UoM
		if selectUom3 != nil {
			getUomResponse, err := s.uomsSS.Get(context.Background(), &connect.Request[pbUoms.GetRequest]{
				Msg: &pbUoms.GetRequest{
					Select: selectUom3,
				},
			})
			if err == nil {
				existUom = getUomResponse.Msg.GetUom()
			}
		}
		UomChan3 <- existUom
	}()

	return LegalEntityRelationships{
		Country: <-countryChan,
		UoM1:    <-UomChan1,
		UoM2:    <-UomChan2,
		UoM3:    <-UomChan3,
	}
}

// For legalentities.addresses.
func (s *ServiceServer) GetAddressList(ctx context.Context, legalEntityID string, legalEntity *pbLegalEntities.LegalEntity,
) (*pbLegalEntities.LegalEntity, error) {
	sql := "SELECT label, address_id, main_address, ownership_status, status FROM core.legalentities_addresses WHERE legalentity_id = $1"

	rows, err := s.db.Query(ctx, sql, legalEntityID)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}

	addressInfos := []AddressInfo{}
	for rows.Next() {
		var (
			label           pgtype.Text
			addressID       pgtype.Text
			mainAddress     pgtype.Bool
			ownershipStatus pgtype.Int2
			status          pgtype.Int2
		)
		err := rows.Scan(&label, &addressID, &mainAddress, &ownershipStatus, &status)
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}
		addressInfos = append(addressInfos, AddressInfo{
			Label:           label.String,
			AddressID:       addressID.String,
			MainAddress:     mainAddress.Bool,
			OwnershipStatus: ownershipStatus.Int16,
			Status:          status.Int16,
		})
	}

	legalEntity.Addresses = &pbAddresses.LabeledAddressList{
		List: []*pbAddresses.LabeledAddress{},
	}

	for _, v := range addressInfos {
		getAddressRes, err := s.addressesSS.Get(ctx, &connect.Request[pbAddresses.GetRequest]{
			Msg: &pbAddresses.GetRequest{
				Id: v.AddressID,
			},
		})
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}

		mainAddress := v.MainAddress
		ownershipStatus := pbAddresses.OwnershipStatus(v.OwnershipStatus)
		legalEntity.Addresses.List = append(legalEntity.Addresses.List, &pbAddresses.LabeledAddress{
			Label:           v.Label,
			Address:         getAddressRes.Msg.GetAddress(),
			MainAddress:     &mainAddress,
			OwnershipStatus: &ownershipStatus,
			Status:          pbKYC.Status(v.Status),
		})
	}

	return legalEntity, nil
}

// For legalentities.contacts
func (s *ServiceServer) GetContactList(ctx context.Context, legalEntityID string, legalEntity *pbLegalEntities.LegalEntity,
) (*pbLegalEntities.LegalEntity, error) {
	sql := "SELECT label, contact_id, main_contact, status FROM core.legalentities_contacts WHERE legalentity_id = $1"

	rows, err := s.db.Query(ctx, sql, legalEntityID)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}

	contactInfos := []ContactInfo{}
	for rows.Next() {
		var (
			label       pgtype.Text
			contactID   pgtype.Text
			mainContact pgtype.Bool
			status      pgtype.Int2
		)
		err := rows.Scan(&label, &contactID, &mainContact, &status)
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}
		contactInfos = append(contactInfos, ContactInfo{
			Label:       label.String,
			ContactID:   contactID.String,
			MainContact: mainContact.Bool,
			Status:      status.Int16,
		})
	}

	legalEntity.Contacts = &pbContacts.LabeledContactList{
		List: []*pbContacts.LabeledContact{},
	}

	for _, v := range contactInfos {
		getContactRes, err := s.contactsSS.Get(ctx, &connect.Request[pbContacts.GetRequest]{
			Msg: &pbContacts.GetRequest{
				Id: v.ContactID,
			},
		})
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}

		mainContact := v.MainContact
		legalEntity.Contacts.List = append(legalEntity.Contacts.List, &pbContacts.LabeledContact{
			Label:       v.Label,
			Contact:     getContactRes.Msg.GetContact(),
			MainContact: &mainContact,
			Status:      pbKYC.Status(v.Status),
		})
	}

	return legalEntity, nil
}

func (s *ServiceServer) appendAddressesContacts(ctx context.Context, req *pbLegalEntities.GetListRequest,
	legalEntity *pbLegalEntities.LegalEntity,
) (*pbLegalEntities.LegalEntity, error) {
	var err error

	// if with_addresses == TRUE, then append addresses to response
	if req.Addresses != nil {
		legalEntity, err = s.GetAddressList(ctx, legalEntity.Id, legalEntity)
		if err != nil {
			return nil, err
		}
	}

	// if with_contacts == TRUE, then append contacts to response
	if req.Contacts != nil {
		legalEntity, err = s.GetContactList(ctx, legalEntity.Id, legalEntity)
		if err != nil {
			return nil, err
		}
	}

	return legalEntity, nil
}
