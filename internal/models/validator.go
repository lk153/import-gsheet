package models

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	internalValidator "github.com/lk153/import-gsheet/internal/validator"
)

// Default validator for the models which supports null types.
var validate = validator.New()

func init() {
	// register all sql.Null* types to use the ValidateValuer CustomTypeFunc
	validate.RegisterCustomTypeFunc(ValidateValuer,
		sql.NullString{},
		sql.NullInt64{},
		sql.NullBool{},
		sql.NullFloat64{})

	// return the db tag name instead of field name
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("db"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	internalValidator.RegisterCustomValidation(validate)
}

// ValidateValuer implements validator.CustomTypeFunc
func ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}

	return nil
}
