-- Add new schema named "core"
CREATE SCHEMA "core";

CREATE TABLE core.changelogs (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	table_name varchar NOT NULL,
	field_name varchar NOT NULL,
	type smallint NOT NULL, -- 1:INSERT, 2:UPDATE, 3:DELETE
	old_value varchar,
	new_value varchar,
	timestamp timestamp NOT NULL DEFAULT now()
);

CREATE TABLE core.uoms (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL, -- type + symbol form the Human-Readable Key (HRK): must be unique in table
	symbol varchar NOT NULL, -- type + symbol form the Human-Readable Key (HRK): must be unique in table
	name varchar,
	icon varchar,
	managed_decimals smallint NOT NULL DEFAULT 8,
	displayed_decimals smallint NOT NULL DEFAULT 3,
	reporting_unit boolean NOT NULL DEFAULT false,
	status smallint NOT NULL DEFAULT 0
);

-- fiats are uoms of type 1
-- fields for fiats are inheritated from uoms and only fiats-specific fields must be specified in this table
CREATE TABLE core.fiats (
	id uuid PRIMARY KEY NOT NULL, -- same id as uoms.id
	iso4217_num varchar
);

-- TO-DO: to be completed
CREATE TABLE core.orgs (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	name varchar NOT NULL, -- Human-Readable Key (HRK): must be unique in table
	type smallint NOT NULL DEFAULT 0,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.addresses (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL DEFAULT 0,
    country_id uuid NOT NULL,
	building varchar,
	floor varchar,
	unit varchar,
	street_num varchar,
	street_name varchar,
	district varchar,
	locality varchar,
	zip_code varchar,
	region varchar,
	state varchar,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.contacts (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL DEFAULT 0,
	value varchar NOT NULL,
	status smallint NOT NULL DEFAULT 0
);

-- TO-DO: to be completed
CREATE TABLE core.legalentities (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	name varchar NOT NULL, -- Human-Readable Key (HRK): must be unique in table
	type smallint NOT NULL DEFAULT 0,
	incorporation_country_id uuid NOT NULL,
	incorporation_locality varchar,
	business_registration_no varchar,
	business_registration_alt_no varchar,
	business_registration_valid_until timestamp,
	tax_id varchar,
	currency1_id uuid NOT NULL,
	currency2_id uuid,
	currency3_id uuid,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.legalentities_addresses (
	legalentity_id uuid NOT NULL,
	label varchar NOT NULL,
    address_id uuid,
    status smallint NOT NULL DEFAULT 1, -- Relationship status
    status_timestamp timestamp NOT NULL DEFAULT now(),
    last_modification timestamp NOT NULL DEFAULT now(),
	PRIMARY KEY (legalentity_id, label)
);

CREATE TABLE core.legalentities_contacts (
	legalentity_id uuid NOT NULL,
	label varchar NOT NULL,
    contact_id uuid,
    status smallint NOT NULL DEFAULT 1, -- Relationship status
    status_timestamp timestamp NOT NULL DEFAULT now(),
    last_modification timestamp NOT NULL DEFAULT now(),
	PRIMARY KEY (legalentity_id, label)
);

-- TO-DO: to be completed
CREATE TABLE core.ledgers (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	name varchar NOT NULL, -- Human-Readable Key (HRK): must be unique in table
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.users (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	login varchar NOT NULL, -- Human-Readable Key (HRK): must be unique in table
	type smallint NOT NULL DEFAULT 1,
	screen_name varchar,
	avatar varchar,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.countries (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    code varchar NOT NULL, -- Human-Readable Key (HRK): must be unique in table
	name varchar,
	icon varchar,
    iso3166_a3 varchar,
    iso3166_num varchar,
    internet_cctld varchar,
	region varchar,
	sub_region varchar,
	intermediate_region varchar,
	intermediate_region_code varchar,
	region_code varchar,
	sub_region_code varchar,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.countries_uoms (
	country_id uuid NOT NULL,
	uom_id uuid NOT NULL,
    status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (country_id, uom_id)
);

-- TO-DO: to be completed
-- cryptos are uoms of type 2
-- fields for cryptos are inheritated from uoms and only cryptos-specific fields must be specified in this table
CREATE TABLE core.cryptos (
	id uuid PRIMARY KEY NOT NULL,
	crypto_type smallint NOT NULL DEFAULT 0
);

-- TO-DO: to be completed
CREATE TABLE core.blockchains (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name varchar NOT NULL, -- Human-Readable Key (HRK): must be unique in table
    type smallint NOT NULL DEFAULT 0,
	icon varchar,
    evm boolean NOT NULL DEFAULT false,
    status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.cryptocategories (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	name varchar NOT NULL, -- Human-Readable Key (HRK): must be unique in table
	icon varchar,
	parent_id uuid,
    status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.blockchains_cryptos (
	blockchain_id uuid NOT NULL,
	crypto_id uuid NOT NULL,
    status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (blockchain_id, crypto_id)
);

CREATE TABLE core.cryptocategories_cryptos (
	cryptocategory_id uuid NOT NULL,
	crypto_id uuid NOT NULL,
    status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (cryptocategory_id, crypto_id)
);

CREATE TABLE core.fsproviders (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL, -- type + name form the Human-Readable Key (HRK): must be unique in table
	name varchar NOT NULL, -- type + name form the Human-Readable Key (HRK): must be unique in table
	icon varchar,
    status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.recipients (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    legalentity_id uuid, -- Human-Readable Key (HRK): (legalentity_id+user_id+label) must be unique in table
    user_id uuid, -- Human-Readable Key (HRK): (legalentity_id+user_id+label) must be unique in table
    label varchar NOT NULL, -- Human-Readable Key (HRK): (legalentity_id+user_id+label) must be unique in table
    type smallint NOT NULL,
    org_id uuid,
    status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.recipients_shared (
	recipient_id uuid NOT NULL, -- Recipient.id
	user_id uuid NOT NULL,
	valid_from timestamp NOT NULL,
	balance bool NOT NULL DEFAULT true,
	history bool NOT NULL DEFAULT false,
	history_from timestamp,
	history_to timestamp,
	deposit bool NOT NULL DEFAULT false,
	deposit_min_amount decimal,
	deposit_max_amount decimal,
	withdrawal bool NOT NULL DEFAULT false,
	withdrawal_max_amount decimal,
	withdrawal_max_percentage decimal,
	transfer bool NOT NULL DEFAULT false,
	transfer_max_amount decimal,
	transfer_max_percentage decimal,
	exchange bool NOT NULL DEFAULT false,
	exchange_max_amount decimal,
	exchange_max_percentage decimal,
    status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (recipient_id, user_id, valid_from)
);

-- dvfiataccounts are recipients of type 1
-- fields for dvfiataccounts are inheritated from recipients and only dvfiataccounts-specific fields must be specified in this table
CREATE TABLE core.dvfiataccounts (
	id uuid PRIMARY KEY NOT NULL, -- Recipient.id
	bankbranch_id uuid NOT NULL,
	bankaccount_type smallint NOT NULL DEFAULT 0,
	currency_id uuid, -- NULL if multi-currency account
	pan varchar NOT NULL,
	masked_pan varchar,
	bban varchar,
	iban varchar,
	external_id varchar
);

-- TO-DO: to be completed
-- dvcryptowallets are recipients of type 2
-- fields for dvcryptowallets are inheritated from recipients and only dvcryptowallets-specific fields must be specified in this table
CREATE TABLE core.dvcryptowallets (
	id uuid PRIMARY KEY NOT NULL, -- Recipient.id
	wallet_type smallint NOT NULL,
	blockchain_id uuid NOT NULL,
	address varchar NOT NULL
);

-- TO-DO: to be completed
-- dvsubaccounts are recipients of type 3
-- fields for dvsubaccounts are inheritated from recipients and only dvsubaccounts-specific fields must be specified in this table
CREATE TABLE core.dvsubaccounts (
	id uuid PRIMARY KEY NOT NULL, -- Recipient.id
	subaccount_type smallint NOT NULL,
	address varchar NOT NULL
);

-- TO-DO: to be completed
-- dvbots are recipients of type 4
-- fields for dvbots are inheritated from recipients and only dvbots-specific fields must be specified in this table
CREATE TABLE core.dvbots (
	id uuid PRIMARY KEY NOT NULL, -- Recipient.id
	bot_type smallint NOT NULL,
	default_params_name varchar,
	bot_state smallint NOT NULL DEFAULT 1,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.dvbots_params (
	bot_id uuid NOT NULL,
	key varchar NOT NULL, -- Key name may start with section/sub-section name with '.' used as a separator
	value varchar NOT NULL,
	status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (bot_id, key)
);

CREATE TABLE core.dvbots_params_default (
	name varchar NOT NULL, -- Template name
	key varchar NOT NULL, -- Key name may start with section/sub-section name with '.' used as a separator
	value varchar NOT NULL,
	status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (name, key)
);

-- TO-DO: to be completed
-- cexaccounts are recipients of type 8
-- fields for cexaccounts are inheritated from recipients and only cexaccounts-specific fields must be specified in this table
CREATE TABLE core.cexaccounts (
	id uuid PRIMARY KEY NOT NULL, -- Recipient.id
	fsprovider_id uuid NOT NULL
);

-- TO-DO: to be completed
-- defiwallets are recipients of type 32
-- fields for defiwallets are inheritated from recipients and only defiwallets-specific fields must be specified in this table
CREATE TABLE core.defiwallets (
	id uuid PRIMARY KEY NOT NULL, -- Recipient.id
	blockchain_id uuid NOT NULL,
	address varchar NOT NULL
);

CREATE TABLE core.ibans (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	country_id uuid NOT NULL, -- Human-Readable Key (HRK): (country_id+valid_from) must be unique in table
	valid_from timestamp NOT NULL DEFAULT now(), -- Human-Readable Key (HRK): (country_id+valid_from) must be unique in table
	algorithm smallint NOT NULL,
	format varchar NOT NULL,
	weights varchar,
	modulo varchar,
	complement varchar,
	method varchar,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.banks (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	name varchar NOT NULL, -- Human-Readable Key (HRK): name must be unique in table
	type smallint NOT NULL DEFAULT 0,
	bic varchar NOT NULL,
	bank_code varchar NOT NULL,
	address_id uuid,
	contact1_id uuid,
	contact2_id uuid,
	contact3_id uuid,
	openbanking_support bool NOT NULL DEFAULT false, -- if true
	parent_id uuid,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.bankbranches (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	bank_id uuid NOT NULL, -- Human-Readable Key (HRK): (bank_id+branch_code) must be unique in table
	branch_code varchar NOT NULL, -- Human-Readable Key (HRK): (bank_id+branch_code) must be unique in table
	type smallint NOT NULL DEFAULT 0, -- same potential values as banks.type (but value can be different from related bank)
	name varchar NOT NULL,
	address_id uuid,
	contact1_id uuid,
	contact2_id uuid,
	contact3_id uuid,
	status smallint NOT NULL DEFAULT 0 -- same potential values as banks.status (but value can be different from related bank)
);

-- TO-DO: to be completed
-- bankaccounts are recipients of type 16
-- fields for bankaccounts are inheritated from recipients and only bankaccounts-specific fields must be specified in this table
CREATE TABLE core.bankaccounts (
	id uuid PRIMARY KEY NOT NULL, -- Recipient.id
	bankbranch_id uuid NOT NULL,
	bankaccount_type smallint NOT NULL DEFAULT 0,
	currency_id uuid, -- NULL if multi-currency account
	pan varchar NOT NULL,
	masked_pan varchar,
	bban varchar,
	iban varchar,
	external_id varchar
);

CREATE TABLE core.userprefs_default (
	key varchar NOT NULL,
	valid_from timestamp NOT NULL DEFAULT now(),
	value varchar NOT NULL,
	status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (key, valid_from)
);

CREATE TABLE core.userprefs (
	user_id uuid NOT NULL,
	key varchar NOT NULL,
	value varchar NOT NULL,
	status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (user_id, key)
);

CREATE TABLE core.uservaults (
	user_id uuid NOT NULL,
	key varchar NOT NULL,
	value_type smallint NOT NULL,
	data bytea NOT NULL,
	status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (user_id, key)
);

CREATE TABLE core.userids (
	user_id uuid PRIMARY KEY NOT NULL,
	credential_id uuid,
	physique_id uuid,
	liveliness_id uuid,
	social_id uuid,
    status smallint NOT NULL DEFAULT 1 -- KYC Status
);

CREATE TABLE core.users_addresses (
	user_id uuid NOT NULL,
	label varchar NOT NULL,
    address_id uuid,
	main_address bool NOT NULL DEFAULT true,
	ownership_status smallint NOT NULL DEFAULT 0, -- 0:UNSPECIFIED, 1:OWNED, 2:RENTED, 3:HOSTED, 255:OTHER
    status smallint NOT NULL DEFAULT 1, -- KYC Status
	PRIMARY KEY (user_id, label)
);

CREATE TABLE core.users_contacts (
	user_id uuid NOT NULL,
	label varchar NOT NULL,
	contact_id uuid NOT NULL,
	main_contact bool NOT NULL DEFAULT false,
    status smallint NOT NULL DEFAULT 1, -- KYC Status
	PRIMARY KEY (user_id, label)
);

CREATE TABLE core.users_incomes (
	user_id uuid NOT NULL,
	label varchar NOT NULL,
	income_id uuid NOT NULL,
    status smallint NOT NULL DEFAULT 1, -- KYC Status
	PRIMARY KEY (user_id, label)
);

CREATE TABLE core.kyc_credentials (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	photo varchar,
	gender varchar,
	title varchar,
	first_name varchar,
	middle_names varchar,
	last_name varchar,
    birthday timestamp,
	country_of_birth_id uuid,
	country_of_nationality_id uuid,
	status smallint NOT NULL DEFAULT 0 -- KYC Status
);

CREATE TABLE core.kyc_physiques (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	race varchar,
	ethnicity varchar,
	eyes_color varchar,
	hair_color varchar,
	body_shape varchar,
	height varchar,
	weight varchar,
	status smallint NOT NULL DEFAULT 0 -- KYC Status
);

CREATE TABLE core.kyc_liveliness (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	liveliness_video_file varchar,
	liveliness_video_file_type varchar NOT NULL, -- MIME type
	timestamp_video_file varchar,
	timestamp_video_file_type varchar NOT NULL, -- MIME type
	id_ownership_photo_file varchar,
	id_ownership_photo_file_type varchar NOT NULL, -- MIME type
	status smallint NOT NULL DEFAULT 0 -- KYC Status
);

CREATE TABLE core.kyc_socials (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	relationship_status varchar,
	religion varchar,
	social_class varchar,
	profession varchar,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.kyc_incomes (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL DEFAULT 0, -- 1:SALARY, 2:FEES, 3:DIVIDENDS, 4:INVESTMENT, 5:RENT, 6:PENSION, 255:OTHER
    amount_year decimal,
    amount_month decimal,
    amount_week decimal,
    amount_day decimal,
    amount_hour decimal,
	currency_id uuid,
    description varchar,
    employer varchar,
    industry varchar,
    occupation varchar,
    employment_type varchar,
    employment_status varchar,
    employment_start_date varchar,
    company varchar,
    investment_vehicle varchar,
    property_address_id uuid NOT NULL,
    from_country_id uuid NOT NULL,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.documents (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	file varchar NOT NULL,
	file_type varchar NOT NULL, -- MIME type
    file_timestamp timestamp NOT NULL DEFAULT now(),
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.documents_data (
	document_id uuid NOT NULL,
	field varchar NOT NULL,
	value varchar NOT NULL,
	status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (document_id, field)
);

CREATE TABLE core.kyc_proofs (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    section smallint NOT NULL, -- 1:credentials, 2:physiques, 3:socials, 4:residences, 5:contacts, 6:incomes
	record_id uuid NOT NULL, -- credentials.id, physiques.id, socials.id, residences.id, contacts.id, incomes.id
	document_type varchar NOT NULL,
	name varchar NOT NULL,
	document_id uuid NOT NULL,
    status smallint NOT NULL DEFAULT 1 -- KYC Status
);

CREATE TABLE core.countries_kyc_info_requirements (
	country_id uuid NOT NULL,
	validity timestamp NOT NULL DEFAULT now(),
	section smallint NOT NULL, -- 1:credentials, 2:physiques, 3:socials, 4:residences, 5:contacts, 6:incomes
	field varchar NOT NULL,
	required bool NOT NULL DEFAULT false,
	optional bool NOT NULL DEFAULT false,
	PRIMARY KEY (country_id, validity, section, field)
);

CREATE TABLE core.countries_kyc_proofs_requirements (
	country_id uuid NOT NULL,
	validity timestamp NOT NULL DEFAULT now(),
	section smallint NOT NULL, -- 1:credentials, 2:physiques, 3:socials, 4:residences, 5:contacts, 6:incomes
	nb_proofs smallint NOT NULL DEFAULT 1,
	accepted_document_type_1 varchar,
	accepted_document_type_2 varchar,
	accepted_document_type_3 varchar,
	accepted_document_type_4 varchar,
	accepted_document_type_5 varchar,
	PRIMARY KEY (country_id, validity, section)
);

CREATE TABLE core.tradingpairs (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	symbol varchar NOT NULL,
	quantity_uom_id uuid NOT NULL,
	quantity_decimals smallint NOT NULL,
	price_uom_id uuid NOT NULL,
	price_decimals smallint NOT NULL,
	volume_decimals smallint NOT NULL,
	status smallint NOT NULL
);

CREATE TABLE core.markets (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	symbol varchar NOT NULL,
	type smallint NOT NULL,
	tradingpair_id uuid NOT NULL,
	algorithm smallint NOT NULL,
	price_type smallint NOT NULL,
	tick_size decimal,
	state smallint NOT NULL,
	status smallint NOT NULL
);

CREATE TABLE core.countries_markets (
	country_id uuid NOT NULL,
	region varchar,
	market_id uuid NOT NULL,
	valid_from timestamp NOT NULL DEFAULT now(),
    status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (country_id, region, market_id, valid_from)
);

CREATE TABLE core.markets_config (
	country_id uuid NOT NULL,
	region varchar,
	market_id uuid NOT NULL,
	valid_from timestamp NOT NULL DEFAULT now(),
	parameter varchar NOT NULL,
	value varchar,
    status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (country_id, region, market_id, valid_from, parameter)
);

CREATE TABLE core.datasources (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL, -- type + name form the Human-Readable Key (HRK): must be unique in table
	name varchar NOT NULL, -- type + name form the Human-Readable Key (HRK): must be unique in table
	fsprovider_id uuid NOT NULL,
	icon varchar,
    status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.prices (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	source_id uuid NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table
	market_id uuid NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table
	type smallint NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table, 1:LTP (Last Trade Price), 2:MTM (Mark-to-Market), 3:MID, 4:BID, 5:ASK, 6:VWAP (Volume-Weighted Average Price), 7:TWAP (Time-Weighted Average Price), 8:ARRIVAL
	timestamp timestamp NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table
	price decimal NOT NULL,
    status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.ohlcvt (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	source_id uuid NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table
	market_id uuid NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table
	price_type smallint NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table, 1:LTP (Last Trade Price), 2:MTM (Mark-to-Market), 3:MID, 4:BID, 5:ASK, 6:VWAP (Volume-Weighted Average Price), 7:TWAP (Time-Weighted Average Price), 8:ARRIVAL
	timestamp timestamp NOT NULL, -- source_id + market_id + type + timestamp form the Human-Readable Key (HRK): must be unique in table
	open decimal,
	high decimal,
	low decimal,
	close decimal,
	volume_in_quantity_uom decimal,
	volume_in_price_uom decimal,
	trades integer,
    status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.authgroups (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	name varchar NOT NULL,
	status smallint NOT NULL DEFAULT 0
);

CREATE TABLE core.transactions (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL DEFAULT 1,
	source_id uuid NOT NULL,
	legalentity_id uuid NOT NULL,
	ledger_id uuid NOT NULL,
	posting_date timestamp NOT NULL,
	accounting_period varchar NOT NULL,
	accounting_document varchar NOT NULL,
	total_amount_in_transaction_currency decimal NOT NULL DEFAULT 0.0,
	transaction_currency_id uuid NOT NULL,
	total_amount_in_legalentity_currency1 decimal NOT NULL DEFAULT 0.0,
	legalentity_currency1_id uuid NOT NULL,
	price_in_legalentity_currency1 decimal NOT NULL DEFAULT 0.0,
	price_id_legalentity_currency1 uuid NOT NULL,
	total_amount_in_legalentity_currency2 decimal,
	legalentity_currency2_id uuid,
	price_in_legalentity_currency2 decimal,
	price_id_legalentity_currency2 uuid,
	total_amount_in_legalentity_currency3 decimal,
	legalentity_currency3_id uuid,
	price_in_legalentity_currency3 decimal,
	price_id_legalentity_currency3 uuid,
	reference varchar,
	prupose varchar,
	user_id uuid NOT NULL,
	authgroup_id uuid,
	org_id uuid,
	status smallint NOT NULL DEFAULT 1
);

CREATE TABLE core.transactions_alt (
	transaction_id uuid NOT NULL, -- id + alt form the Primary Key
	alt smallint NOT NULL, -- id + alt form the Primary Key
	amount decimal NOT NULL DEFAULT 0.0,
	currency_id uuid NOT NULL,
	price decimal NOT NULL DEFAULT 0.0,
	price_id uuid,
	status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (transaction_id, alt)
);

CREATE TABLE core.transactionitems (
	transaction_id uuid NOT NULL, -- transaction_id + item_no form the Human-Readable Key
	item_no integer NOT NULL, -- transaction_id + item_no form the Human-Readable Key
	type smallint NOT NULL DEFAULT 1, -- 1:DEBIT, 2:CREDIT
	recipient_id uuid,
	financial_account varchar,
	amount_in_transaction_currency decimal NOT NULL DEFAULT 0.0,
	transaction_currency_id uuid NOT NULL,
	amount_in_legalentity_currency1 decimal NOT NULL DEFAULT 0.0,
	legalentity_currency1_id uuid NOT NULL,
	price_in_legalentity_currency1 decimal NOT NULL DEFAULT 0.0,
	price_id_legalentity_currency1 uuid NOT NULL,
	amount_in_legalentity_currency2 decimal NOT NULL DEFAULT 0.0,
	legalentity_currency2_id uuid NOT NULL,
	price_in_legalentity_currency2 decimal NOT NULL DEFAULT 0.0,
	price_id_legalentity_currency2 uuid NOT NULL,
	amount_in_legalentity_currency3 decimal NOT NULL DEFAULT 0.0,
	legalentity_currency3_id uuid NOT NULL,
	price_in_legalentity_currency3 decimal NOT NULL DEFAULT 0.0,
	price_id_legalentity_currency3 uuid NOT NULL,
	reference varchar,
	legalentity_id_offset uuid,
	user_id_offset uuid,
	org_id_offset uuid,
	recipient_id_offset uuid,
	transaction_id_offset uuid,
	item_no_offset integer,
	status smallint NOT NULL DEFAULT 0,
	PRIMARY KEY (transaction_id, item_no)
);

CREATE TABLE core.transactionitems_alt (
	transaction_id uuid NOT NULL, -- id + item_no + alt form the Primary Key
	item_no integer NOT NULL, -- id + item_no + alt form the Primary Key
	alt smallint NOT NULL, -- id + item_no + alt form the Primary Key
	amount decimal NOT NULL DEFAULT 0.0,
	currency_id uuid NOT NULL,
	price decimal NOT NULL DEFAULT 0.0,
	price_id uuid,
	status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (transaction_id, item_no, alt)
);

CREATE TABLE core.balances (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	type smallint NOT NULL, -- 1:ACTUAL, 2:UNREALIZED
	recipient_id uuid NOT NULL, -- type + recipient_id + timestamp form the Human-Readable Key
	timestamp timestamp NOT NULL, -- type + recipient_id + timestamp form the Human-Readable Key
	transaction_id uuid NOT NULL,
	item_no integer NOT NULL,
	amount_in_transaction_currency decimal NOT NULL DEFAULT 0.0,
	transaction_currency_id uuid NOT NULL,
	amount_in_legalentity_currency1 decimal NOT NULL DEFAULT 0.0,
	legalentity_currency1_id uuid NOT NULL,
	price_in_legalentity_currency1 decimal NOT NULL DEFAULT 0.0,
	price_id_legalentity_currency1 uuid NOT NULL,
	amount_in_legalentity_currency2 decimal NOT NULL DEFAULT 0.0,
	legalentity_currency2_id uuid NOT NULL,
	price_in_legalentity_currency2 decimal NOT NULL DEFAULT 0.0,
	price_id_legalentity_currency2 uuid NOT NULL,
	amount_in_legalentity_currency3 decimal NOT NULL DEFAULT 0.0,
	legalentity_currency3_id uuid NOT NULL,
	price_in_legalentity_currency3 decimal NOT NULL DEFAULT 0.0,
	price_id_legalentity_currency3 uuid NOT NULL,
	status smallint NOT NULL DEFAULT 1
);

CREATE TABLE core.balances_alt (
	balance_id uuid NOT NULL,
	alt smallint NOT NULL,
	amount decimal NOT NULL DEFAULT 0.0,
	currency_id uuid NOT NULL,
	price decimal NOT NULL DEFAULT 0.0,
	price_id uuid,
	status smallint NOT NULL DEFAULT 1,
	PRIMARY KEY (balance_id, alt)
);
