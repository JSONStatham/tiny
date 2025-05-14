package httpserver

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
)

func validateWithTrans(s any) validator.ValidationErrorsTranslations {
	en := en.New()
	uni = ut.New(en, en)

	validate = validator.New()

	// Use JSON tag name for error messages
	validate.RegisterTagNameFunc(withTagName)

	trans, ok := uni.GetTranslator("en")
	if ok {
		en_translations.RegisterDefaultTranslations(validate, trans)
	}

	if err := validate.Struct(s); err != nil {
		errs := err.(validator.ValidationErrors)
		return errs.Translate(trans)
	}

	return nil
}

func withTagName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	// skip if tag key says it should be ignored
	if name == "-" {
		return ""
	}
	return name
}
