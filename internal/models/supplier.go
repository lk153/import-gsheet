package models

import (
	"database/sql"
)

type Supplier struct {
	Id                       int64           `db:"id"`
	CompanyName              string          `db:"company_name"`
	AlternateCompanyName     sql.NullString  `db:"alternate_company_name"`
	Country                  string          `db:"country"`
	City                     sql.NullString  `db:"city"`
	Entity                   string          `db:"entity"`
	LocationRegion           sql.NullString  `db:"location_region"`
	ContactNumber            string          `db:"contact_number"`
	LegalPerson              sql.NullString  `db:"legal_person"`
	ContactPerson            string          `db:"contact_person"`
	SocialNetworkId          sql.NullString  `db:"social_network_id"`
	SocialNetworkType        sql.NullString  `db:"social_network_type"`
	Ranking                  sql.NullString  `db:"ranking"`
	PassedVetting            sql.NullString  `db:"passed_vetting"`
	VettingInfoUrl           sql.NullString  `db:"vetting_info_url"`
	ClassificationID         sql.NullInt64   `db:"classification_id"`
	NumberOfEmployeesRangeID sql.NullInt64   `db:"number_of_employees_range_id"`
	Status                   sql.NullString  `db:"status"`
	LegalPersonId            sql.NullString  `db:"legal_person_id"`
	IsLegacy                 bool            `db:"is_legacy"`
	SupplierDetails          *SupplierDetail `db:"details"`
	CreatedAt                sql.NullTime    `db:"created_at"`
	CreatedBy                sql.NullString  `db:"created_by"`
	UpdatedAt                sql.NullTime    `db:"updated_at"`
	UpdatedBy                sql.NullString  `db:"updated_by"`
	DeletedAt                sql.NullTime    `db:"deleted_at"`
	DeletedBy                sql.NullString  `db:"deleted_by"`
}
