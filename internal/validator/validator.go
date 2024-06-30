package validator

import (
	"context"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

const (
	DuplicateSupplier     = "Duplicated Supplier"
	IncorrectInformation  = "Incorrect information"
	SupplierWithoutOrders = "Supplier without orders"
)

var DeleteChangeReasons = []string{
	DuplicateSupplier,
	IncorrectInformation,
	SupplierWithoutOrders,
}

var conform = modifiers.New()

var DefaultValidator = &defaultValidator{}

// defaultValidator implements gin binding.StructValidator
// similar to binding.defaultValidator but add some custom validators and use mold to transform data
type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *defaultValidator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		// Added this line to transform data before validation. Only need to transform when obj is a pointer
		if err := transform(obj); err != nil {
			return err
		}
		return v.ValidateStruct(reflect.ValueOf(obj).Elem().Interface())
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(binding.SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

// validateStruct receives struct type
func (v *defaultValidator) validateStruct(obj any) error {
	v.lazyinit()
	return v.validate.Struct(obj)
}

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://pkg.go.dev/github.com/go-playground/validator/v10
func (v *defaultValidator) Engine() any {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
		/*	Need this part so the validator return the json tag name
			instead of fieldName in the error response  */
		v.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		RegisterCustomValidation(v.validate)
	})
}

func RegisterCustomValidation(v *validator.Validate) {
	err := v.RegisterValidation("customUrl", customURLValidator)
	if err != nil {
		log.Err(err).Msg("Error while registering custom URL validator")
		panic(err)
	}

	err = v.RegisterValidation("customEmail", customEmailValidator)
	if err != nil {
		log.Err(err).Msg("Error while registering custom Email validator")
		panic(err)
	}

	err = v.RegisterValidation("customSocialNetworkId", customSocialNetworkValidator)
	if err != nil {
		log.Err(err).Msg("Error while registering custom Social Network Id validator")
		panic(err)
	}

	err = v.RegisterValidation("customISO8601", customISO8601Validator)
	if err != nil {
		log.Err(err).Msg("Error while registering custom ISO8601 time format validator")
		panic(err)
	}

	err = v.RegisterValidation("customNotBlank", validators.NotBlank)
	if err != nil {
		log.Err(err).Msg("Error while registering custom Not Blank validator")
		panic(err)
	}

	err = v.RegisterValidation("customDeleteReason", customDeleteReasonValidator)
	if err != nil {
		log.Err(err).Msg("Error while registering custom Not Blank validator")
		panic(err)
	}

	err = v.RegisterValidation("customNoSpace", customNoSpace)
	if err != nil {
		log.Err(err).Msg("Error while registering custom No Space validator")
		panic(err)
	}
}

func transform(obj any) error {
	if conform == nil {
		return nil
	}
	ctx := context.Background()
	value := reflect.ValueOf(reflect.ValueOf(obj).Elem().Interface())
	switch value.Kind() {
	case reflect.Struct:
		return conform.Struct(ctx, obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		for i := 0; i < count; i++ {
			val := value.Index(i)
			switch val.Kind() {
			case reflect.Ptr:
				el := val.Interface()
				if err := conform.Struct(ctx, el); err != nil {
					return err
				}
				value.Index(i).Set(reflect.ValueOf(el))
			case reflect.Struct:
				el := val.Addr().Interface()
				if err := conform.Struct(ctx, el); err != nil {
					return err
				}
				value.Index(i).Set(reflect.ValueOf(el).Elem())
			default:
			}
		}
	default:
		return nil
	}
	return nil
}

func customURLValidator(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	parsedURI, err := url.ParseRequestURI(val)
	if err != nil {
		return false
	}

	scheme := parsedURI.Scheme
	if scheme != "http" && scheme != "https" {
		return false
	}

	return true
}

func customEmailValidator(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	emailRegex := regexp.MustCompile(`^([a-z0-9_+-]+\.?)*[a-z0-9_+-]@([a-z0-9][a-z0-9\-]*\.)+[a-z]{2,}$`)
	return emailRegex.MatchString(val)
}

func customSocialNetworkValidator(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	if len(val) < 6 || len(val) > 20 {
		return false
	}

	return true
}

func customISO8601Validator(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	_, err := time.Parse("2006-01-02", val)
	return err == nil
}

func customDeleteReasonValidator(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	return slices.Contains(DeleteChangeReasons, val)
}

func customNoSpace(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	return !lo.SomeBy([]rune(val), unicode.IsSpace)
}
