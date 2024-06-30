package models

import (
	"database/sql"
)

type BankAccountDetails struct {
	Id                     int64          `db:"id"`
	AccountType            sql.NullString `db:"account_type" validate:"omitempty,oneof=Corporate Personal"`
	AccountHolderName      sql.NullString `db:"account_holder_name" validate:"required_with=AccountType"`
	AccountNumber          sql.NullString `db:"account_number" validate:"required_with=AccountType,gte=0"`
	BankName               sql.NullString `db:"bank_name" validate:"required_with=AccountType"`
	SwiftCode              sql.NullString `db:"swift_code" validate:"omitempty,min=8,max=11,customNoSpace"`
	BankAddress            sql.NullString `db:"bank_address"`
	SupplierId             int64          `db:"supplier_id"`
	SupplierCompanyAddress sql.NullString `db:"supplier_company_address"`
	CreatedAt              sql.NullTime   `db:"created_at"`
	UpdatedAt              sql.NullTime   `db:"updated_at"`
	DeletedAt              sql.NullTime   `db:"deleted_at"`
}
