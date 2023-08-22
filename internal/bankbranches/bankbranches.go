package bankbranches

import (
	"context"
	"fmt"
	"strings"
	"sync"

	pbAddresses "davensi.com/core/gen/addresses"
	pbBankBranches "davensi.com/core/gen/bankbranches"
	pbBankBranchesConnect "davensi.com/core/gen/bankbranches/bankbranchesconnect"
	pbBanks "davensi.com/core/gen/banks"
	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	"davensi.com/core/internal/addresses"
	"davensi.com/core/internal/banks"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/contacts"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	_package          = "bankbranches"
	_entityName       = "Bank Branch"
	_entityNamePlural = "Bank Branches"
)

// For singleton BankBranch export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the BankBranches API
type ServiceServer struct {
	Repo BankBranchRepository
	pbBankBranchesConnect.UnimplementedServiceHandler
	db          *pgxpool.Pool
	banksSS     *banks.ServiceServer
	addressesSS *addresses.ServiceServer
	contactsSS  *contacts.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:        *NewBankRepository(db),
		db:          db,
		banksSS:     banks.GetSingletonServiceServer(db),
		addressesSS: addresses.GetSingletonServiceServer(db),
		contactsSS:  contacts.GetSingletonServiceServer(db),
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Create(ctx context.Context, req *connect.Request[pbBankBranches.CreateRequest],
) (*connect.Response[pbBankBranches.CreateResponse], error) {
	var (
		addressUUID       string
		addressCreationFn func(tx pgx.Tx) (*pbAddresses.Address, error)

		contact1UUID, contact2UUID, contact3UUID                   = uuid.NewString(), uuid.NewString(), uuid.NewString()
		contact1CreationFn, contact2CreationFn, contact3CreationFn func(tx pgx.Tx) (*pbContacts.Contact, error)

		genErr *common.ErrWithCode
	)

	// Optional Address field
	if req.Msg.Address != nil {
		addressUUID = uuid.NewString()
		addressCreationFn, genErr = s.addressesSS.GenCreateFunc(req.Msg.Address, addressUUID)
		if genErr != nil {
			log.Error().Err(genErr.Err)
			return connect.NewResponse(&pbBankBranches.CreateResponse{
				Response: &pbBankBranches.CreateResponse_Error{
					Error: &pbCommon.Error{
						Code:    genErr.Code,
						Package: _package,
						Text:    genErr.Err.Error(),
					},
				},
			}), genErr.Err
		}
	}

	// Optional Contact1,2,3 fields
	contact1CreationFn, contact2CreationFn, contact3CreationFn, genErr = s.genContactCreateFn(
		req.Msg, contact1UUID, contact2UUID, contact3UUID)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbBankBranches.CreateResponse{
			Response: &pbBankBranches.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	bankBranchCreationFn, genErr := s.GenCreateFunc(req.Msg, addressUUID, contact1UUID, contact2UUID, contact3UUID)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbBankBranches.CreateResponse{
			Response: &pbBankBranches.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	var newBankBranch *pbBankBranches.BankBranch
	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if req.Msg.Address != nil {
			_, errWriteAddress := addressCreationFn(tx)
			if errWriteAddress != nil {
				return nil
			}
		}

		if req.Msg.Contact1 != nil {
			_, errWriteContact1 := contact1CreationFn(tx)
			if errWriteContact1 != nil {
				return nil
			}
		}

		if req.Msg.Contact2 != nil {
			_, errWriteContact2 := contact2CreationFn(tx)
			if errWriteContact2 != nil {
				return nil
			}
		}

		if req.Msg.Contact3 != nil {
			_, errWriteContact3 := contact3CreationFn(tx)
			if errWriteContact3 != nil {
				return nil
			}
		}

		excutedBankBranch, errWriteBankBranch := bankBranchCreationFn(tx)
		if errWriteBankBranch != nil {
			return nil
		}
		newBankBranch = excutedBankBranch

		return errWriteBankBranch
	}); errExcute != nil {
		commonErrCreate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_entityName,
			errExcute.Error(),
		)

		log.Error().Err(commonErrCreate.Err)
		return connect.NewResponse(&pbBankBranches.CreateResponse{
			Response: &pbBankBranches.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErrCreate.Code,
					Package: _package,
					Text:    commonErrCreate.Err.Error(),
				},
			},
		}), commonErrCreate.Err
	}

	log.Info().Msgf(
		"%s created successfully with id = %s",
		_entityName, newBankBranch.GetId(),
	)

	return connect.NewResponse(&pbBankBranches.CreateResponse{
		Response: &pbBankBranches.CreateResponse_Bankbranch{
			Bankbranch: newBankBranch,
		},
	}), nil
}

func (s *ServiceServer) GenCreateFunc(req *pbBankBranches.CreateRequest, addressUUID, contact1UUID, contact2UUID, contact3UUID string) (
	func(tx pgx.Tx) (*pbBankBranches.BankBranch, error), *common.ErrWithCode,
) {
	errGenFn := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR, "creating", _entityName, "")

	if validateErr := s.validateCreate(req); validateErr != nil {
		return nil, validateErr
	}

	qb, errInsert := s.Repo.QbInsert(req, addressUUID, contact1UUID, contact2UUID, contact3UUID)
	if errInsert != nil {
		return nil, errGenFn.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errInsert.Error())
	}

	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlStr + "'")

	return func(tx pgx.Tx) (*pbBankBranches.BankBranch, error) {
		executedBankBranch, errWriteBankBranch := common.TxWrite[pbBankBranches.BankBranch](
			context.Background(),
			tx,
			sqlStr,
			args,
			ScanRow,
		)

		if errWriteBankBranch != nil {
			return nil, errWriteBankBranch
		}

		return executedBankBranch, nil
	}, nil
}

func (s *ServiceServer) genContactCreateFn(req *pbBankBranches.CreateRequest, contact1UUID, contact2UUID, contact3UUID string) (
	fn1 func(tx pgx.Tx) (*pbContacts.Contact, error),
	fn2 func(tx pgx.Tx) (*pbContacts.Contact, error),
	fn3 func(tx pgx.Tx) (*pbContacts.Contact, error),
	genError *common.ErrWithCode,
) {
	var (
		contact1CreationFn func(tx pgx.Tx) (*pbContacts.Contact, error)
		contact2CreationFn func(tx pgx.Tx) (*pbContacts.Contact, error)
		contact3CreationFn func(tx pgx.Tx) (*pbContacts.Contact, error)
		genErr             *common.ErrWithCode
	)

	// Optional Contact1 field
	if req.Contact1 != nil {
		contact1CreationFn, genErr = s.contactsSS.GenCreateFunc(req.Contact1, contact1UUID)
		if genErr != nil {
			log.Error().Err(genErr.Err)
			return nil, nil, nil, genErr
		}
	}

	// Optional Contact2 field
	if req.Contact2 != nil {
		contact2CreationFn, genErr = s.contactsSS.GenCreateFunc(req.Contact2, contact2UUID)
		if genErr != nil {
			log.Error().Err(genErr.Err)
			return nil, nil, nil, genErr
		}
	}

	// Optional Contact3 field
	if req.Contact3 != nil {
		contact3CreationFn, genErr = s.contactsSS.GenCreateFunc(req.Contact3, contact3UUID)
		if genErr != nil {
			log.Error().Err(genErr.Err)
			return nil, nil, nil, genErr
		}
	}

	return contact1CreationFn, contact2CreationFn, contact3CreationFn, nil
}

func (s *ServiceServer) Update(
	ctx context.Context, req *connect.Request[pbBankBranches.UpdateRequest],
) (*connect.Response[pbBankBranches.UpdateResponse], error) {
	// Validation
	if errQueryUpdate := s.validateUpdateQuery(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbBankBranches.UpdateResponse{
			Response: &pbBankBranches.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	// check if old Bank Branch exists. Also to get IDs for address, contact1, contact2, contact3
	oldBankBranch, err := s.getOldBankBranchToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbBankBranches.UpdateResponse{
			Response: &pbBankBranches.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    "update failed: could not get old Bank Branch to update",
				},
			},
		}), err
	}

	// LOGIC: if oldBankBranch.Address.Id == "" then INSERT, else UPDATE
	var (
		addressUUID       = uuid.NewString()
		addressCreationFn func(tx pgx.Tx) (*pbAddresses.Address, error)
		addressUpdateFn   func(tx pgx.Tx) (*pbAddresses.Address, error)
		genErr            *common.ErrWithCode
	)
	if req.Msg.GetAddress() != nil {
		if oldBankBranch.Msg.GetBankbranch().Address.Id == "" {
			// CREATE ADDRESS
			addressCreationFn, genErr = s.addressesSS.GenCreateFunc(&pbAddresses.CreateRequest{
				Type:       req.Msg.Address.Type,
				Country:    req.Msg.Address.Country,
				Building:   req.Msg.Address.Building,
				Floor:      req.Msg.Address.Floor,
				Unit:       req.Msg.Address.Unit,
				StreetNum:  req.Msg.Address.StreetNum,
				StreetName: req.Msg.Address.StreetName,
				District:   req.Msg.Address.District,
				Locality:   req.Msg.Address.Locality,
				ZipCode:    req.Msg.Address.ZipCode,
				Region:     req.Msg.Address.Region,
				State:      req.Msg.Address.State,
				Status:     req.Msg.Address.Status,
			}, addressUUID)
		} else {
			// UPDATE ADDRESS
			addressUUID = oldBankBranch.Msg.GetBankbranch().Address.Id // used later to update bankbranches.address_id to same existing value
			addressUpdateFn, _, genErr = s.addressesSS.GenUpdateFunc(&pbAddresses.UpdateRequest{
				Id:         oldBankBranch.Msg.GetBankbranch().Address.Id,
				Type:       req.Msg.Address.Type,
				Country:    req.Msg.Address.Country,
				Building:   req.Msg.Address.Building,
				Floor:      req.Msg.Address.Floor,
				Unit:       req.Msg.Address.Unit,
				StreetNum:  req.Msg.Address.StreetNum,
				StreetName: req.Msg.Address.StreetName,
				District:   req.Msg.Address.District,
				Locality:   req.Msg.Address.Locality,
				ZipCode:    req.Msg.Address.ZipCode,
				Region:     req.Msg.Address.Region,
				State:      req.Msg.Address.State,
				Status:     req.Msg.Address.Status,
			})
		}
		if genErr != nil {
			log.Error().Err(genErr.Err)
			return connect.NewResponse(&pbBankBranches.UpdateResponse{
				Response: &pbBankBranches.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    genErr.Code,
						Package: _package,
						Text:    genErr.Err.Error(),
					},
				},
			}), genErr.Err
		}
	}

	// For contact1, contact2, contact3
	var (
		contact1UUID = uuid.NewString()
		contact2UUID = uuid.NewString()
		contact3UUID = uuid.NewString()
	)
	createFuncs, updateFuncs, contactsUUIDs, genErr := s.genUpsertContactFuncs(req, oldBankBranch, contact1UUID, contact2UUID, contact3UUID)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbBankBranches.UpdateResponse{
			Response: &pbBankBranches.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	// generate update SQL for main entity (Bank Branch)
	// Update addressID, contact123ID: update to same existing id if already linked, else write new
	qb, genSQLError := s.Repo.QbUpdate(req.Msg, addressUUID, contactsUUIDs[0], contactsUUIDs[1], contactsUUIDs[2])
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbBankBranches.UpdateResponse{
			Response: &pbBankBranches.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	qb.SetReturnFields("*")
	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	// Executing update and saving response
	var (
		updatedBankBranch *pbBankBranches.BankBranch
		errScan           error
	)
	log.Info().Msg("Executing SQL '" + sqlstr + "'")
	err = crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// UPSERT ADDRESS
		err := executeUpsertAddress(req, oldBankBranch, addressCreationFn, addressUpdateFn, tx)
		if err != nil {
			return err
		}

		// UPSERT CONTACT1, CONTACT2, CONTACT3
		for i := 0; i < 3; i++ {
			err := executeUpsertContact(req, oldBankBranch, createFuncs[i], updateFuncs[i], tx, i)
			if err != nil {
				return err
			}
		}

		// UPDATE MAIN ENTITY
		row, err := tx.Query(ctx, sqlstr, sqlArgs...)
		if err != nil {
			return err
		}

		if row.Next() {
			updatedBankBranch, errScan = ScanRow(row)
			if errScan != nil {
				log.Error().Err(err).Msgf("unable to update %s with identifier = '%s'", _entityName, sel)
				return errScan
			}
			log.Info().Msgf("%s updated successfully with id = %s",
				_entityName, updatedBankBranch.GetId())
			row.Close()
		}
		return nil
	})
	if err != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)], "updating", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbBankBranches.UpdateResponse{
			Response: &pbBankBranches.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	return connect.NewResponse(&pbBankBranches.UpdateResponse{
		Response: &pbBankBranches.UpdateResponse_Bankbranch{
			Bankbranch: updatedBankBranch,
		},
	}), nil
}

func (s *ServiceServer) getOldBankBranchToUpdate(req *pbBankBranches.UpdateRequest) (*connect.Response[pbBankBranches.GetResponse], error) {
	var getBankBranchRequest *pbBankBranches.GetRequest
	switch req.GetSelect().GetSelect().(type) {
	case *pbBankBranches.Select_ById:
		getBankBranchRequest = &pbBankBranches.GetRequest{
			Select: &pbBankBranches.Select{
				Select: &pbBankBranches.Select_ById{
					ById: req.GetSelect().GetById(),
				},
			},
		}
	case *pbBankBranches.Select_ByBankBranchCode:
		switch req.GetSelect().GetByBankBranchCode().GetBank().GetSelect().(type) {
		case *pbBanks.Select_ById:
			getBankBranchRequest = &pbBankBranches.GetRequest{
				Select: &pbBankBranches.Select{
					Select: &pbBankBranches.Select_ByBankBranchCode{
						ByBankBranchCode: &pbBankBranches.BankBranchCode{
							Bank: &pbBanks.Select{
								Select: &pbBanks.Select_ById{
									ById: req.GetSelect().GetByBankBranchCode().GetBank().GetById(),
								},
							},
							BranchCode: req.GetSelect().GetByBankBranchCode().GetBranchCode(),
						},
					},
				},
			}
		case *pbBanks.Select_ByName:
			getBankBranchRequest = &pbBankBranches.GetRequest{
				Select: &pbBankBranches.Select{
					Select: &pbBankBranches.Select_ByBankBranchCode{
						ByBankBranchCode: &pbBankBranches.BankBranchCode{
							Bank: &pbBanks.Select{
								Select: &pbBanks.Select_ByName{
									ByName: req.GetSelect().GetByBankBranchCode().GetBank().GetByName(),
								},
							},
							BranchCode: req.GetSelect().GetByBankBranchCode().GetBranchCode(),
						},
					},
				},
			}
		}
	}

	getBankBranchRes, err := s.Get(context.Background(), &connect.Request[pbBankBranches.GetRequest]{
		Msg: getBankBranchRequest,
	})
	if err != nil {
		return nil, err
	}

	return getBankBranchRes, nil
}

func (s *ServiceServer) genUpsertContactFuncs(
	req *connect.Request[pbBankBranches.UpdateRequest],
	oldBankBranch *connect.Response[pbBankBranches.GetResponse],
	contact1UUID, contact2UUID, contact3UUID string) (
	createFuncs [3]func(tx pgx.Tx) (*pbContacts.Contact, error),
	updateFuncs [3]func(tx pgx.Tx) (*pbContacts.Contact, error),
	contactsUUIDs [3]string, // fixed length arrays used to prevent out-of-range error when calling executeUpsertContact()
	err *common.ErrWithCode,
) {
	contact1CreationFn, contact1UpdateFn, contact1UUID, genErr := s.genOneUpsertContactFunc(
		req.Msg.Contact1, oldBankBranch.Msg.GetBankbranch().Contact1, contact1UUID)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return [3]func(tx pgx.Tx) (*pbContacts.Contact, error){}, [3]func(tx pgx.Tx) (*pbContacts.Contact, error){}, [3]string{}, genErr
	}
	contact2CreationFn, contact2UpdateFn, contact2UUID, genErr := s.genOneUpsertContactFunc(
		req.Msg.Contact2, oldBankBranch.Msg.GetBankbranch().Contact2, contact2UUID)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return [3]func(tx pgx.Tx) (*pbContacts.Contact, error){}, [3]func(tx pgx.Tx) (*pbContacts.Contact, error){}, [3]string{}, genErr
	}
	contact3CreationFn, contact3UpdateFn, contact3UUID, genErr := s.genOneUpsertContactFunc(
		req.Msg.Contact3, oldBankBranch.Msg.GetBankbranch().Contact3, contact3UUID)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return [3]func(tx pgx.Tx) (*pbContacts.Contact, error){}, [3]func(tx pgx.Tx) (*pbContacts.Contact, error){}, [3]string{}, genErr
	}

	// For contact1
	createFuncs[0] = contact1CreationFn
	updateFuncs[0] = contact1UpdateFn
	contactsUUIDs[0] = contact1UUID

	createFuncs[1] = contact2CreationFn
	updateFuncs[1] = contact2UpdateFn
	contactsUUIDs[1] = contact2UUID

	createFuncs[2] = contact3CreationFn
	updateFuncs[2] = contact3UpdateFn
	contactsUUIDs[2] = contact3UUID

	return createFuncs, updateFuncs, contactsUUIDs, nil
}

func (s *ServiceServer) genOneUpsertContactFunc(
	req *pbContacts.UpdateContact,
	bankBranchContact *pbContacts.Contact,
	contactUUIDArg string,
) (
	contactCreationFn func(tx pgx.Tx) (*pbContacts.Contact, error),
	contactUpdateFn func(tx pgx.Tx) (*pbContacts.Contact, error),
	contactUUIDReturn string,
	genErr *common.ErrWithCode,
) {
	if req != nil {
		if bankBranchContact.Id == "" {
			// CREATE CONTACT
			contactCreationFn, genErr = s.contactsSS.GenCreateFunc(&pbContacts.CreateRequest{
				Type:   req.GetType(),
				Value:  req.GetValue(),
				Status: req.Status,
			}, contactUUIDArg)
			if genErr != nil {
				log.Error().Err(genErr.Err)
				return nil, nil, "", genErr
			}
			// Note which UUID to return: newly generated for INSERT or existing value from oldBankBranch (bankBranchContact.Id) for UPDATE
			return contactCreationFn, nil, contactUUIDArg, nil
		} else {
			// UPDATE CONTACT2
			contactUpdateFn, _, genErr = s.contactsSS.GenUpdateFunc(&pbContacts.UpdateRequest{
				Id:     bankBranchContact.Id,
				Type:   req.Type,
				Value:  req.Value,
				Status: req.Status,
			})
			if genErr != nil {
				log.Error().Err(genErr.Err)
				return nil, nil, "", genErr
			}
			return nil, contactUpdateFn, bankBranchContact.Id, nil
		}
	}

	return nil, nil, "", nil
}

func executeUpsertAddress(req *connect.Request[pbBankBranches.UpdateRequest],
	oldBankBranch *connect.Response[pbBankBranches.GetResponse],
	createFunc func(tx pgx.Tx) (*pbAddresses.Address, error),
	updateFunc func(tx pgx.Tx) (*pbAddresses.Address, error),
	tx pgx.Tx,
) error {
	if req.Msg.Address != nil {
		var errWriteAddress error
		if oldBankBranch.Msg.GetBankbranch().Address.Id != "" {
			// UPDATE ADDRESS
			_, errWriteAddress = updateFunc(tx)
		} else {
			// INSERT ADDRESS
			_, errWriteAddress = createFunc(tx)
		}
		if errWriteAddress != nil {
			return errWriteAddress
		}
	} else {
		return nil
	}

	return nil
}

func executeUpsertContact(req *connect.Request[pbBankBranches.UpdateRequest],
	oldBankBranch *connect.Response[pbBankBranches.GetResponse],
	createFunc func(tx pgx.Tx) (*pbContacts.Contact, error),
	updateFunc func(tx pgx.Tx) (*pbContacts.Contact, error),
	tx pgx.Tx,
	contactIndex int, // contactIndex are used to dictate which IF block is ran: Contact1 or Contact2 or Contact3
) error {
	if req.Msg.Contact1 != nil && contactIndex == 0 {
		var errWriteContact1 error
		if oldBankBranch.Msg.GetBankbranch().Contact1.Id == "" {
			// INSERT ADDRESS IF NULL
			_, errWriteContact1 = createFunc(tx)
		} else {
			// UPDATE ADDRESS IF NOT NULL
			_, errWriteContact1 = updateFunc(tx)
		}
		if errWriteContact1 != nil {
			return errWriteContact1
		}
	}

	if req.Msg.Contact2 != nil && contactIndex == 1 {
		var errWriteContact2 error
		if oldBankBranch.Msg.GetBankbranch().Contact2.Id == "" {
			// INSERT ADDRESS IF NULL
			_, errWriteContact2 = createFunc(tx)
		} else {
			// UPDATE ADDRESS IF NOT NULL
			_, errWriteContact2 = updateFunc(tx)
		}
		if errWriteContact2 != nil {
			return errWriteContact2
		}
	}

	if req.Msg.Contact3 != nil && contactIndex == 2 {
		var errWriteContact3 error
		if oldBankBranch.Msg.GetBankbranch().Contact3.Id == "" {
			// INSERT ADDRESS IF NULL
			_, errWriteContact3 = createFunc(tx)
		} else {
			// UPDATE ADDRESS IF NOT NULL
			_, errWriteContact3 = updateFunc(tx)
		}
		if errWriteContact3 != nil {
			return errWriteContact3
		}
	}

	return nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbBankBranches.DeleteRequest],
) (*connect.Response[pbBankBranches.DeleteResponse], error) {
	var updateRequest pbBankBranches.UpdateRequest
	switch req.Msg.GetSelect().GetSelect().(type) {
	case *pbBankBranches.Select_ById:
		updateRequest.Select = &pbBankBranches.Select{
			Select: &pbBankBranches.Select_ById{
				ById: req.Msg.GetSelect().GetById(),
			},
		}
		updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
	case *pbBankBranches.Select_ByBankBranchCode:
		switch req.Msg.GetSelect().GetByBankBranchCode().GetBank().GetSelect().(type) {
		case *pbBanks.Select_ById:
			updateRequest.Select = &pbBankBranches.Select{
				Select: &pbBankBranches.Select_ByBankBranchCode{
					ByBankBranchCode: &pbBankBranches.BankBranchCode{
						Bank: &pbBanks.Select{
							Select: &pbBanks.Select_ById{
								ById: req.Msg.GetSelect().GetByBankBranchCode().GetBank().GetById(),
							},
						},
						BranchCode: req.Msg.GetSelect().GetByBankBranchCode().GetBranchCode(),
					},
				},
			}
			updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
		case *pbBanks.Select_ByName:
			updateRequest.Select = &pbBankBranches.Select{
				Select: &pbBankBranches.Select_ByBankBranchCode{
					ByBankBranchCode: &pbBankBranches.BankBranchCode{
						Bank: &pbBanks.Select{
							Select: &pbBanks.Select_ByName{
								ByName: req.Msg.GetSelect().GetByBankBranchCode().GetBank().GetByName(),
							},
						},
						BranchCode: req.Msg.GetSelect().GetByBankBranchCode().GetBranchCode(),
					},
				},
			}
			updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
		}
	}

	updateRes, err := s.Update(ctx, connect.NewRequest(&updateRequest))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbBankBranches.DeleteResponse{
			Response: &pbBankBranches.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, "id = "+updateRes.Msg.GetBankbranch().Id)
	return connect.NewResponse(&pbBankBranches.DeleteResponse{
		Response: &pbBankBranches.DeleteResponse_Bankbranch{
			Bankbranch: updateRes.Msg.GetBankbranch(),
		},
	}), nil
}

func (s *ServiceServer) Get(ctx context.Context, req *connect.Request[pbBankBranches.GetRequest],
) (*connect.Response[pbBankBranches.GetResponse], error) {
	commonErr := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, "fetching", _entityName, "")

	if errQueryGet := ValidateSelect(req.Msg.GetSelect(), "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbBankBranches.GetResponse{
			Response: &pbBankBranches.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	var (
		qbBank     = s.banksSS.Repo.QbGetList(&pbBanks.GetListRequest{})
		qbAddress  = s.addressesSS.Repo.QbGetList(&pbAddresses.GetListRequest{})
		qbContact1 = s.contactsSS.Repo.QbGetList(&pbContacts.GetListRequest{})
		qbContact2 = s.contactsSS.Repo.QbGetList(&pbContacts.GetListRequest{})
		qbContact3 = s.contactsSS.Repo.QbGetList(&pbContacts.GetListRequest{})
	)

	var (
		filterBankStr, filterBankArgs         = qbBank.Filters.GenerateSQL()
		filterAddressStr, filterAddressArgs   = qbAddress.Filters.GenerateSQL()
		filterContact1Str, filterContact1Args = qbContact1.Filters.GenerateSQL()
		filterContact2Str, filterContact2Args = qbContact2.Filters.GenerateSQL()
		filterContact3Str, filterContact3Args = qbContact3.Filters.GenerateSQL()
	)

	qb := s.Repo.QbGetOne(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankbranches.bank_id = banks.id", qbBank.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON (bankbranches.address_id = addresses.id AND addresses.status = 1)", qbAddress.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON (bankbranches.contact1_id = contacts.id AND contacts.status = 1)", qbContact1.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON (bankbranches.contact2_id = contacts.id AND contacts.status = 1)", qbContact2.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON (bankbranches.contact3_id = contacts.id AND contacts.status = 1)", qbContact3.TableName)).
		Select(strings.Join(qbBank.SelectFields, ", ")).
		Select(strings.Join(qbAddress.SelectFields, ", ")).
		Select(strings.Join(qbContact1.SelectFields, ", ")).
		Select(strings.Join(qbContact2.SelectFields, ", ")).
		Select(strings.Join(qbContact3.SelectFields, ", ")).
		Where(filterBankStr, filterBankArgs...).
		Where(filterAddressStr, filterAddressArgs...).
		Where(filterContact1Str, filterContact1Args...).
		Where(filterContact2Str, filterContact2Args...).
		Where(filterContact3Str, filterContact3Args...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	// Transform sqlstr to give aliases to core.uoms because there are multiple left-joins on that table
	replaceCount := 4
	sqlstr = strings.Replace(sqlstr, "contacts", "contact1", replaceCount)
	sqlstr = strings.Replace(sqlstr, "contacts", "contact2", replaceCount)
	sqlstr = strings.Replace(sqlstr, "contacts", "contact3", replaceCount)
	sqlstr = strings.Replace(sqlstr, "core.contacts ON (bankbranches.contact1_id = contacts.id AND contacts.status = 1)",
		"core.contacts contact1 ON (bankbranches.contact1_id = contact1.id AND contact1.status = 1)", 1)
	sqlstr = strings.Replace(sqlstr, "core.contacts ON (bankbranches.contact2_id = contacts.id AND contacts.status = 1)",
		"core.contacts contact2 ON (bankbranches.contact2_id = contact2.id AND contact2.status = 1)", 1)
	sqlstr = strings.Replace(sqlstr, "core.contacts ON (bankbranches.contact3_id = contacts.id AND contacts.status = 1)",
		"core.contacts contact3 ON (bankbranches.contact3_id = contact3.id AND contact3.status = 1)", 1)

	log.Info().Msg("Executing SQL '" + sqlstr + "'")
	rows, errQueryRows := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errQueryRows != nil {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).UpdateMessage(errQueryRows.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankBranches.GetResponse{
			Response: &pbBankBranches.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}
	defer rows.Close()

	if !rows.Next() {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND).UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankBranches.GetResponse{
			Response: &pbBankBranches.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	bankBranch, errScanRow := s.Repo.ScanWithRelationship(rows)
	if errScanRow != nil {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).UpdateMessage(errScanRow.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankBranches.GetResponse{
			Response: &pbBankBranches.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	if rows.Next() {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND).UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankBranches.GetResponse{
			Response: &pbBankBranches.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbBankBranches.GetResponse{
		Response: &pbBankBranches.GetResponse_Bankbranch{
			Bankbranch: bankBranch,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbBankBranches.GetListRequest],
	res *connect.ServerStream[pbBankBranches.GetListResponse],
) error {
	if req.Msg.Bank == nil {
		req.Msg.Bank = &pbBanks.GetListRequest{}
	}

	if req.Msg.Address == nil {
		req.Msg.Address = &pbAddresses.GetListRequest{}
	}

	if req.Msg.Contact1 == nil {
		req.Msg.Contact1 = &pbContacts.GetListRequest{}
	}

	if req.Msg.Contact2 == nil {
		req.Msg.Contact2 = &pbContacts.GetListRequest{}
	}

	if req.Msg.Contact3 == nil {
		req.Msg.Contact3 = &pbContacts.GetListRequest{}
	}

	var (
		qbBank     = s.banksSS.Repo.QbGetList(req.Msg.Bank)
		qbAddress  = s.addressesSS.Repo.QbGetList(req.Msg.Address)
		qbContact1 = s.contactsSS.Repo.QbGetList(req.Msg.Contact1)
		qbContact2 = s.contactsSS.Repo.QbGetList(req.Msg.Contact2)
		qbContact3 = s.contactsSS.Repo.QbGetList(req.Msg.Contact3)

		filterBankStr, filterBankArgs         = qbBank.Filters.GenerateSQL()
		filterAddressStr, filterAddressArgs   = qbAddress.Filters.GenerateSQL()
		filterContact1Str, filterContact1Args = qbContact1.Filters.GenerateSQL()
		filterContact2Str, filterContact2Args = qbContact2.Filters.GenerateSQL()
		filterContact3Str, filterContact3Args = qbContact3.Filters.GenerateSQL()
	)

	// Assigning different aliases to core.contacts, because there are multiple JOINs on that same table
	filterContact1Str = strings.ReplaceAll(filterContact1Str, "contacts", "contact1")
	filterContact2Str = strings.ReplaceAll(filterContact2Str, "contacts", "contact2")
	filterContact3Str = strings.ReplaceAll(filterContact3Str, "contacts", "contact3")

	qb := s.Repo.QbGetList(req.Msg)
	qb.Join(fmt.Sprintf("LEFT JOIN %s ON bankbranches.bank_id = banks.id", qbBank.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankbranches.address_id = addresses.id", qbAddress.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankbranches.contact1_id = contacts.id", qbContact1.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankbranches.contact2_id = contacts.id", qbContact2.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankbranches.contact3_id = contacts.id", qbContact3.TableName)).
		Select(strings.Join(qbBank.SelectFields, ", ")).
		Select(strings.Join(qbAddress.SelectFields, ", ")).
		Select(strings.Join(qbContact1.SelectFields, ", ")).
		Select(strings.Join(qbContact2.SelectFields, ", ")).
		Select(strings.Join(qbContact3.SelectFields, ", ")).
		Where(filterBankStr, filterBankArgs...).
		Where(filterAddressStr, filterAddressArgs...).
		Where(filterContact1Str, filterContact1Args...).
		Where(filterContact2Str, filterContact2Args...).
		Where(filterContact3Str, filterContact3Args...)

	sqlStr, sqlArgs := generateAndTransformSQL(qb)
	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, sqlArgs...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbBankBranches.GetListResponse{
					Response: &pbBankBranches.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}
	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		bankBranch, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbBankBranches.GetListResponse{
						Response: &pbBankBranches.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbBankBranches.GetListResponse{
			Response: &pbBankBranches.GetListResponse_Bankbranch{
				Bankbranch: bankBranch,
			},
		}); errSend != nil {
			_errSend := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR)], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

// func streamingErr(res *connect.ServerStream[pbBankBranches.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
// 	_errno := errorCode
// 	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
// 	log.Error().Err(err).Msg(_err.Error())

// 	if errSend := res.Send(&pbBankBranches.GetListResponse{
// 		Response: &pbBankBranches.GetListResponse_Error{
// 			Error: &pbCommon.Error{
// 				Code: _errno,
// 				Text: _err.Error() + " (" + err.Error() + ")",
// 			},
// 		},
// 	}); errSend != nil {
// 		_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
// 		_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
// 		log.Error().Err(errSend).Msg(_errSend.Error())
// 		_err = _errSend
// 	}
// 	return _err
// }
