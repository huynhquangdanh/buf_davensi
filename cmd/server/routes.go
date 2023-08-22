package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	pbAddressesConnect "davensi.com/core/gen/addresses/addressesconnect"
	pbAuthGroupsConnect "davensi.com/core/gen/authgroups/authgroupsconnect"
	pbBankAccountsConnect "davensi.com/core/gen/bankaccounts/bankaccountsconnect"
	pbBankBranchesConnect "davensi.com/core/gen/bankbranches/bankbranchesconnect"
	pbBanksConnect "davensi.com/core/gen/banks/banksconnect"
	pbBlockchainsConnect "davensi.com/core/gen/blockchains/blockchainsconnect"
	pbCexaccountsConnect "davensi.com/core/gen/cexaccounts/cexaccountsconnect"
	pbContactsConnect "davensi.com/core/gen/contacts/contactsconnect"
	pbCountriesConnect "davensi.com/core/gen/countries/countriesconnect"
	pbCredentialsConnect "davensi.com/core/gen/credentials/credentialsconnect"
	pbCryptosConnect "davensi.com/core/gen/cryptos/cryptosconnect"
	pbDataSourcesConnect "davensi.com/core/gen/datasources/datasourcesconnect"
	pbDefiwalletsConnect "davensi.com/core/gen/defiwallets/defiwalletsconnect"
	pbDocumentsConnect "davensi.com/core/gen/documents/documentsconnect"
	pbDVBotsConnect "davensi.com/core/gen/dvbots/dvbotsconnect"
	pbDVSubAccountsConnect "davensi.com/core/gen/dvsubaccounts/dvsubaccountsconnect"
	pbFiatsConnect "davensi.com/core/gen/fiats/fiatsconnect"
	pbFSProvidersConnect "davensi.com/core/gen/fsproviders/fsprovidersconnect"
	pbIbansConnect "davensi.com/core/gen/ibans/ibansconnect"
	pbIncomesConnect "davensi.com/core/gen/incomes/incomesconnect"
	pbLedgersConnect "davensi.com/core/gen/ledgers/ledgersconnect"
	pbLegalEntitiesConnect "davensi.com/core/gen/legalentities/legalentitiesconnect"
	pbLivelinessConnect "davensi.com/core/gen/liveliness/livelinessconnect"
	pbMarketsConnect "davensi.com/core/gen/markets/marketsconnect"
	pbOhlcvtConnect "davensi.com/core/gen/ohlcvt/ohlcvtconnect"
	pbOrgsConnect "davensi.com/core/gen/orgs/orgsconnect"
	pbPhysiquesConnect "davensi.com/core/gen/physiques/physiquesconnect"
	pbPricesConnect "davensi.com/core/gen/prices/pricesconnect"
	pbProofsConnect "davensi.com/core/gen/proofs/proofsconnect"
	pbRecipientsConnect "davensi.com/core/gen/recipients/recipientsconnect"
	pbSocialsConnect "davensi.com/core/gen/socials/socialsconnect"
	pbTradingPairsConnect "davensi.com/core/gen/tradingpairs/tradingpairsconnect"
	pbUoMsConnect "davensi.com/core/gen/uoms/uomsconnect"
	pbUserIDsConnect "davensi.com/core/gen/userids/useridsconnect"
	pbUserPrefsConnect "davensi.com/core/gen/userprefs/userprefsconnect"
	pbUsersConnect "davensi.com/core/gen/users/usersconnect"
	pbUserVaultsConnect "davensi.com/core/gen/uservaults/uservaultsconnect"

	pbAddresses "davensi.com/core/internal/addresses"
	pbAuthGroups "davensi.com/core/internal/authgroups"
	pbBankAccounts "davensi.com/core/internal/bankaccounts"
	pbBankBranches "davensi.com/core/internal/bankbranches"
	pbBanks "davensi.com/core/internal/banks"
	pbBlockchains "davensi.com/core/internal/blockchains"
	pbCexaccounts "davensi.com/core/internal/cexaccounts"
	pbContacts "davensi.com/core/internal/contacts"
	pbCountries "davensi.com/core/internal/countries"
	pbCredentials "davensi.com/core/internal/credentials"
	pbCryptos "davensi.com/core/internal/cryptos"
	pbDataSources "davensi.com/core/internal/datasources"
	pbDefiwallets "davensi.com/core/internal/defiwallets"
	pbDocuments "davensi.com/core/internal/documents"
	pbDVBots "davensi.com/core/internal/dvbots"
	pbDvSubAccounts "davensi.com/core/internal/dvsubaccounts"
	pbFiats "davensi.com/core/internal/fiats"
	pbFSProviders "davensi.com/core/internal/fsproviders"
	pbIbans "davensi.com/core/internal/ibans"
	pbIncomes "davensi.com/core/internal/incomes"
	pbLedgers "davensi.com/core/internal/ledgers"
	pbLegalEntities "davensi.com/core/internal/legalentities"
	pbLiveliness "davensi.com/core/internal/livelinesses"
	pbMarkets "davensi.com/core/internal/markets"
	pbOhlcvt "davensi.com/core/internal/ohlcvt"
	pbOrgs "davensi.com/core/internal/orgs"
	pbPhysiques "davensi.com/core/internal/physiques"
	pbPrices "davensi.com/core/internal/prices"
	pbProofs "davensi.com/core/internal/proofs"
	pbRecipients "davensi.com/core/internal/recipients"
	pbSocials "davensi.com/core/internal/socials"
	pbTradingPairs "davensi.com/core/internal/tradingpairs"
	pbUoMs "davensi.com/core/internal/uoms"
	pbUserIDs "davensi.com/core/internal/userids"
	pbUserPrefs "davensi.com/core/internal/userprefs"
	pbUsers "davensi.com/core/internal/users"
	pbUserVaults "davensi.com/core/internal/uservaults"
)

//nolint:funlen
func routes(conn *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	routesKYC(conn, mux)
	routesUoms(conn, mux)

	path, handler := pbAddressesConnect.NewServiceHandler(pbAddresses.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbAuthGroupsConnect.NewServiceHandler(pbAuthGroups.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbBankBranchesConnect.NewServiceHandler(pbBankBranches.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbBankAccountsConnect.NewServiceHandler(pbBankAccounts.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbBanksConnect.NewServiceHandler(pbBanks.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbBlockchainsConnect.NewServiceHandler(pbBlockchains.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbCexaccountsConnect.NewServiceHandler(pbCexaccounts.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbContactsConnect.NewServiceHandler(pbContacts.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbCountriesConnect.NewServiceHandler(pbCountries.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbDataSourcesConnect.NewServiceHandler(pbDataSources.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbDefiwalletsConnect.NewServiceHandler(pbDefiwallets.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbDVBotsConnect.NewServiceHandler(pbDVBots.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbDVSubAccountsConnect.NewServiceHandler(pbDvSubAccounts.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbFSProvidersConnect.NewServiceHandler(pbFSProviders.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbDocumentsConnect.NewServiceHandler(pbDocuments.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbIbansConnect.NewServiceHandler(pbIbans.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbLedgersConnect.NewServiceHandler(pbLedgers.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbLegalEntitiesConnect.NewServiceHandler(pbLegalEntities.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbMarketsConnect.NewServiceHandler(pbMarkets.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbOhlcvtConnect.NewServiceHandler(pbOhlcvt.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbOrgsConnect.NewServiceHandler(pbOrgs.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbPricesConnect.NewServiceHandler(pbPrices.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbRecipientsConnect.NewServiceHandler(pbRecipients.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbUserIDsConnect.NewServiceHandler(pbUserIDs.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbUserPrefsConnect.NewServiceHandler(pbUserPrefs.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbUsersConnect.NewServiceHandler(pbUsers.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbUserVaultsConnect.NewServiceHandler(pbUserVaults.NewServiceServer(conn))
	mux.Handle(path, handler)

	return mux
}

func routesKYC(conn *pgxpool.Pool, mux *http.ServeMux) {
	path, handler := pbCredentialsConnect.NewServiceHandler(pbCredentials.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbIncomesConnect.NewServiceHandler(pbIncomes.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbLivelinessConnect.NewServiceHandler(pbLiveliness.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbPhysiquesConnect.NewServiceHandler(pbPhysiques.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbProofsConnect.NewServiceHandler(pbProofs.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbSocialsConnect.NewServiceHandler(pbSocials.NewServiceServer(conn))
	mux.Handle(path, handler)
}

func routesUoms(conn *pgxpool.Pool, mux *http.ServeMux) {
	path, handler := pbUoMsConnect.NewServiceHandler(pbUoMs.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbTradingPairsConnect.NewServiceHandler(pbTradingPairs.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbCryptosConnect.NewServiceHandler(pbCryptos.NewServiceServer(conn))
	mux.Handle(path, handler)

	path, handler = pbFiatsConnect.NewServiceHandler(pbFiats.NewServiceServer(conn))
	mux.Handle(path, handler)
}
