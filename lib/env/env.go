package env

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/lk153/import-gsheet/lib/configs"
)

var (
	// errInvalidValue returned when the value passed to unmarshal is nil or not a pointer to a struct.
	errInvalidValue = errors.New("value must be a non-nil pointer to a struct")

	// errUnsupportedType returned when a field with tag "env" is unsupported.
	errUnsupportedType = errors.New("field is an unsupported type")

	// errUnexportedField returned when a field is not exported.
	errUnexportedField = errors.New("field must be exported")

	// errMissingEnvName returned when a field doesn't define the envName struct tag
	errMissingEnvName = errors.New("envName field must be defined as a struct tag")

	// errMissingValue returned when a field is tagged with the "mandatory" tag as true, and does not contain a value
	errMissingValue = errors.New("value must be a non-nil pointer to a struct")
)

// NVBaseEnv is a struct containing the default env variables that all services will have attached to it
type NVBaseEnv struct {
	NvEnv         string `envName:"NV_ENV" defaultValue:"dev"`
	NvSystemId    string `envName:"NV_SYSTEM_ID" defaultValue:"global"`
	NvServiceName string `envName:"NV_SERVICE_NAME" defaultValue:"default-service-name"`
}

// An example of the struct definition can be found at NVBaseEnv
func Init(c configs.Config, env interface{}) error {
	return unmarshal(c, env)
}

func unmarshal(c configs.Config, env interface{}) error {
	rv := reflect.ValueOf(env)

	// Throws an error if the env type isn't a pointer, or is an empty pointer
	// An point is enforced to ensure that only 1 instance is created; otherwise you might end up with multiple instances of the
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errInvalidValue
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errInvalidValue
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		valueField := rv.Field(i)
		switch valueField.Kind() {
		case reflect.Struct:
			if !valueField.Addr().CanInterface() {
				// throw an error immediately if the nested struct can't be used as an interface
				return errInvalidValue
			}

			nestedStruct := valueField.Addr().Interface()
			if err := unmarshal(c, nestedStruct); err != nil {
				return err
			}

			// the struct will be set and initialized; there isn't a need to re-set it
			continue
		}

		typeField := t.Field(i)
		envName := typeField.Tag.Get("envName")
		defaultValue := typeField.Tag.Get("defaultValue")
		mandatoryValue := typeField.Tag.Get("mandatory")

		// env tag is mandatory
		if envName == "" {
			return errMissingEnvName
		}

		// If the field value can't be changed, throw an error
		if !valueField.CanSet() {
			return errUnexportedField
		}

		envValue := c.GetString(envName)

		// Set default value if any
		if envValue == "" && defaultValue != "" {
			envValue = defaultValue
		}

		if mandatoryValue == "true" && envValue == "" {
			return fmt.Errorf("%s is required, but not set. %w", envName, errMissingValue)
		}

		err := set(typeField.Type, valueField, envValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// set uses reflection and sets the value to the field based on the data type
// This method assumes that the reflect.type and reflect.value can be modified
func set(t reflect.Type, f reflect.Value, value string) error {
	switch t.Kind() {
	case reflect.Ptr:
		ptr := reflect.New(t.Elem())
		err := set(t.Elem(), ptr.Elem(), value)
		if err != nil {
			return err
		}
		f.Set(ptr)
	case reflect.String:
		f.SetString(value)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		f.SetBool(v)
	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t.PkgPath() == "time" && t.Name() == "Duration" {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}

			f.Set(reflect.ValueOf(duration))
			break
		}

		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		f.SetInt(int64(v))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		f.SetUint(v)
	default:
		return errUnsupportedType
	}

	return nil
}
