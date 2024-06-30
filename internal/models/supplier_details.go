package models

import (
	"database/sql"
)

type SupplierDetail struct {
	Id                         int64            `db:"id"`
	SupplierId                 int64            `db:"supplier_id"`
	BusinessRegistrationNumber sql.NullString   `db:"business_registration_number"`
	PaidUpCapitalRMB           sql.NullInt64    `db:"paid_up_capital_in_rmb"`
	RegisteredBusinessAddress  sql.NullString   `db:"registered_business_address"`
	SupplierAddress            sql.NullString   `db:"supplier_address"`
	DateOfEstablishment        sql.NullTime     `db:"date_of_establishment"`
	EmailAddress               sql.NullString   `db:"email_address"`
	SupplierWebsiteURL         sql.NullString   `db:"supplier_website_url"`
	SupplierType               sql.NullString   `db:"supplier_type"`
	BrandedGoods               int16            `db:"branded_goods"`
	BrandCheckID               sql.NullString   `db:"brand_check_id"`
	OriginSource               string           `db:"origin_source"`
	GMVInRMB                   sql.NullInt64    `db:"gmv_in_rmb"`
	MarginInPercentage         sql.NullInt64    `db:"margin_in_percentage"`
	LastTransactionDate        sql.NullTime     `db:"last_transaction_date"`
	SupplierTierId             sql.NullInt64    `db:"supplier_tier_id"`
	SupplierTierRel            *SupplierTierRel `db:"supplier_tier_rel"`
	LicenseToProduce           sql.NullBool     `db:"license_to_produce"`
	OemAcceptance              sql.NullBool     `db:"oem_acceptance"`
	FactoryProductionLine      sql.NullBool     `db:"factory_production_line"`
	HonestCivilDebtor          sql.NullBool     `db:"honest_civil_debtor"`
	InvoiceUnderNinja          sql.NullBool     `db:"invoice_under_ninja"`
	CreatedAt                  sql.NullTime     `db:"created_at"`
	UpdatedAt                  sql.NullTime     `db:"updated_at"`
	DeletedAt                  sql.NullTime     `db:"deleted_at"`
}
