package usercontacts

import (
	"context"
	"fmt"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbKyc "davensi.com/core/gen/kyc"
	"davensi.com/core/internal/contacts"
)

type ServiceServer struct {
	repo UserContactsRepository
	db   *pgxpool.Pool
}
type ContactsParams struct {
	ID     string
	Type   pbContacts.Type
	Value  string
	Status *pbCommon.Status
}

type ModifyUserContactParams struct {
	UserID      string
	ModifyType  string
	Contact     *ContactsParams
	Status      *pbKyc.Status
	MainContact bool
	Label       string
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewUserContactRepository(db),
		db:   db,
	}
}

var (
	singleSS *ServiceServer
	onceSS   sync.Once
)

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	onceSS.Do(func() {
		singleSS = NewServiceServer(db)
	})
	return singleSS
}

func (s *ServiceServer) checkPkUnique(ctx context.Context, contactID, scanUserID string) ([]string, error) {
	row, err := s.db.Query(
		ctx,
		"SELECT 1 FROM core.users_contacts WHERE user_id = $1 AND contact_id = $2 AND status != $3",
		scanUserID,
		contactID,
		pbCommon.Status_STATUS_TERMINATED,
	)
	if err != nil {
		return nil, err
	}
	if row.Next() {
		return []string{scanUserID, contactID}, fmt.Errorf("record with user_id: %s and contact_id: %s already exists", scanUserID, contactID)
	}
	return []string{scanUserID, contactID}, nil
}
func (s *ServiceServer) ProcessModifyContacts(
	ctx context.Context,
	contactMsg *pbContacts.SetLabeledContact,
	modifyType string, scanUserID string,
	modifyContactsParamsByIDList *[]*ModifyUserContactParams,
	modifyContactsParamsByContactList *[]*ModifyUserContactParams,
	contactSingletonServer contacts.ServiceServer,
) *pbCommon.Error {
	var modifyContactsParams string
	switch contactMsg.GetSelect().(type) {
	case *pbContacts.SetLabeledContact_Id:
		pkes, checkPkUniqueErr := s.checkPkUnique(ctx, contactMsg.GetId(), scanUserID)
		if pkes == nil && checkPkUniqueErr != nil {
			log.Error().Err(checkPkUniqueErr)
			return &pbCommon.Error{
				Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
				Package: _package,
				Text:    checkPkUniqueErr.Error(),
			}
		}
		existContact, existContactErr := contactSingletonServer.Get(ctx, &connect.Request[pbContacts.GetRequest]{
			Msg: &pbContacts.GetRequest{
				Id: pkes[1],
			},
		})
		if existContactErr != nil {
			log.Error().Err(existContactErr)
			return &pbCommon.Error{
				Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
				Package: _package,
				Text:    existContactErr.Error(),
			}
		}

		if modifyType == _contactAddSymbol {
			modifyContactsParams = modifyType
		}
		if modifyType == _contactUpsertSymbol {
			if pkes != nil && checkPkUniqueErr != nil {
				modifyContactsParams = _contactUpdateSymbol
			} else {
				modifyContactsParams = _contactAddSymbol
			}
		}
		*modifyContactsParamsByIDList = append(*modifyContactsParamsByIDList, &ModifyUserContactParams{
			ModifyType:  modifyContactsParams,
			UserID:      pkes[0],
			Status:      contactMsg.GetStatus().Enum(),
			MainContact: contactMsg.GetMainContact(),
			Label:       contactMsg.GetLabel(),
			Contact: &ContactsParams{
				ID:     pkes[1],
				Type:   existContact.Msg.GetContact().GetType(),
				Value:  existContact.Msg.GetContact().GetValue(),
				Status: existContact.Msg.GetContact().GetStatus().Enum(),
			},
		})
	case *pbContacts.SetLabeledContact_Contact:
		if validateCreateContactErr := contactSingletonServer.ValidateCreate(&pbContacts.CreateRequest{
			Value:  contactMsg.GetContact().GetValue(),
			Type:   contactMsg.GetContact().GetType(),
			Status: contactMsg.GetContact().GetStatus().Enum(),
		}); validateCreateContactErr != nil {
			log.Error().Err(validateCreateContactErr.Err)
			return &pbCommon.Error{
				Code:    validateCreateContactErr.Code,
				Package: _package,
				Text:    validateCreateContactErr.Err.Error(),
			}
		}
		*modifyContactsParamsByContactList = append(*modifyContactsParamsByContactList, &ModifyUserContactParams{
			ModifyType:  _contactAddSymbol,
			Status:      contactMsg.GetStatus().Enum(),
			MainContact: contactMsg.GetMainContact(),
			Label:       contactMsg.GetLabel(),
			Contact: &ContactsParams{
				ID:     uuid.NewString(),
				Type:   contactMsg.GetContact().GetType(),
				Value:  contactMsg.GetContact().GetValue(),
				Status: contactMsg.GetContact().GetStatus().Enum(),
			},
		})
	}
	return nil
}
