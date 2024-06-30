package models

import (
	"database/sql"
	"time"
)

type SupplierTier struct {
	Id        int64        `db:"id"`
	Name      string       `db:"name"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

type SupplierTierRel struct {
	Id   *int64  `db:"id"`
	Name *string `db:"name"`
}
