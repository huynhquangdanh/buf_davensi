package blockchains

import (
	"fmt"
	"strings"

	pbBlockchains "davensi.com/core/gen/blockchains"
	pbCommon "davensi.com/core/gen/common"
	pbUoMs "davensi.com/core/gen/uoms"

	"davensi.com/core/internal/uoms"
	"davensi.com/core/internal/util"
)

var uomsRepo = uoms.NewUoMRepository(nil)

func QbUpsertBlockchainCrypto(blockchain *pbBlockchains.Blockchain, selectUoMs *pbUoMs.SelectList) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Upsert, "core.blockchains_cryptos").
		SetInsertField("blockchain_id", "crypto_id", "status")

	for _, selectUoM := range selectUoMs.List {
		_, err := qb.SetInsertValues([]any{
			blockchain.Id,
			selectUoM.GetById(),
			pbCommon.Status_STATUS_ACTIVE,
		})

		if err != nil {
			return qb
		}
	}

	return qb
}

func QbSoftRemoveBlockchainCrypto(blockchain *pbBlockchains.Blockchain, selectUoMs *pbUoMs.SelectList) *util.QueryBuilder {
	uomArgs := []any{}

	for _, selectUoM := range selectUoMs.List {
		uomArgs = append(uomArgs, selectUoM.GetById())
	}

	qb := util.CreateQueryBuilder(util.Update, "core.blockchains_cryptos").
		SetUpdate("status = ?", pbCommon.Status_STATUS_TERMINATED).
		Where("blockchains_cryptos.blockchain_id = ?", blockchain.Id).
		Where(
			fmt.Sprintf("blockchains_cryptos.country_id IN (%s)", strings.Join(strings.Split(strings.Repeat("?", len(selectUoMs.List)), ""), ", ")),
			uomArgs...,
		)

	return qb
}
