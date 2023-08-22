package userincomes

import (
	"context"
	"fmt"
	"sync"

	pbCommon "davensi.com/core/gen/common"
	pbIncomes "davensi.com/core/gen/incomes"
	pbKyc "davensi.com/core/gen/kyc"
	"davensi.com/core/internal/incomes"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type ServiceServer struct {
	repo UserIncomesRepository
	db   *pgxpool.Pool
}
type ModifyUserIncomeParams struct {
	UserID                 string
	ModifyType             string
	Income                 *pbIncomes.Income
	CreateNewIncomeRequest *pbIncomes.CreateRequest
	CreateNewIncomeID      string
	Status                 pbKyc.Status
	Label                  string
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
func (s *ServiceServer) checkIncomePkUnique(ctx context.Context, incomeID, scanUserID string) ([]string, error) {
	row, err := s.db.Query(
		ctx,
		"SELECT 1 FROM core.users_incomes WHERE user_id = $1 AND income_id = $2",
		scanUserID,
		incomeID,
	)
	if err != nil {
		return nil, err
	}
	if row.Next() {
		return []string{scanUserID, incomeID}, fmt.Errorf("record with user_id: %s and contact_id: %s already exists", scanUserID, incomeID)
	}
	return []string{scanUserID, incomeID}, nil
}
func (s *ServiceServer) ProcessModifyIncomes(
	ctx context.Context,
	incomeMsg *pbIncomes.SetLabeledIncome,
	modifyType string, scanUserID string,
	modifyIncomesParamsByIDList *[]*ModifyUserIncomeParams,
	modifyIncomesParamsByContactList *[]*ModifyUserIncomeParams,
	incomeSingletonServer incomes.ServiceServer,
) *pbCommon.Error {
	var modifyContactsParams string
	switch incomeMsg.GetResponse().(type) {
	case *pbIncomes.SetLabeledIncome_Id:
		pkes, checkPkUniqueErr := s.checkIncomePkUnique(ctx, incomeMsg.GetId(), scanUserID)
		if pkes == nil && checkPkUniqueErr != nil {
			log.Error().Err(checkPkUniqueErr)
			return &pbCommon.Error{
				Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
				Package: _package,
				Text:    checkPkUniqueErr.Error(),
			}
		}
		existIncome, existIncomeErr := incomeSingletonServer.GetSpecificIncome(ctx, pkes[1])
		if existIncomeErr != nil {
			log.Error().Err(existIncomeErr)
			return &pbCommon.Error{
				Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
				Package: _package,
				Text:    existIncomeErr.Error(),
			}
		}

		if modifyType == _incomeAddSymbol {
			modifyContactsParams = modifyType
		}
		if modifyType == _incomeUpsertSymbol {
			if pkes != nil && checkPkUniqueErr != nil {
				modifyContactsParams = _incomeUpdateSymbol
			} else {
				modifyContactsParams = _incomeAddSymbol
			}
		}
		*modifyIncomesParamsByIDList = append(*modifyIncomesParamsByIDList, &ModifyUserIncomeParams{
			ModifyType: modifyContactsParams,
			UserID:     pkes[0],
			Status:     incomeMsg.GetStatus(),
			Label:      incomeMsg.GetLabel(),
			Income:     existIncome,
		})
	case *pbIncomes.SetLabeledIncome_Income:
		if validateCreateIncomeErrNo, validateCreateIncomeErr := incomeSingletonServer.ValidateCreation(&pbIncomes.CreateRequest{
			Select: incomeMsg.GetIncome().GetSelect(),
			Status: incomeMsg.GetIncome().GetStatus().Enum(),
		}); validateCreateIncomeErr != nil {
			log.Error().Err(validateCreateIncomeErr)
			return &pbCommon.Error{
				Code:    validateCreateIncomeErrNo,
				Package: _package,
				Text:    validateCreateIncomeErr.Error(),
			}
		}
		*modifyIncomesParamsByContactList = append(*modifyIncomesParamsByContactList, &ModifyUserIncomeParams{
			ModifyType:             _incomeAddSymbol,
			Status:                 incomeMsg.GetStatus(),
			Label:                  incomeMsg.GetLabel(),
			CreateNewIncomeRequest: incomeMsg.GetIncome(),
			CreateNewIncomeID:      uuid.NewString(),
		})
	}
	return nil
}
