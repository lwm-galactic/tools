package validation

import (
	"fmt"
	english "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/en"
	"os"
	"reflect"
	// zh_trans "github.com/go-playground/validator/v10/translations/zh"
)

const (
	maxDescriptionLength = 255
)

// Validator is a custom validator for configs.
type Validator struct {
	val   *validator.Validate
	data  interface{}
	trans ut.Translator
}

// NewValidator creates a new Validator.
func NewValidator(data interface{}) *Validator {
	result := validator.New()

	// independent validators 给这些tag 自定义校验函数示例
	result.RegisterValidation("dir", validateDir)                 // nolint: errcheck // no need
	result.RegisterValidation("file", validateFile)               // nolint: errcheck // no need
	result.RegisterValidation("description", validateDescription) // nolint: errcheck // no need
	result.RegisterValidation("name", validateName)               // nolint: errcheck // no need

	// default translations
	eng := english.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")
	err := en.RegisterDefaultTranslations(result, trans)
	if err != nil {
		panic(err)
	}

	// additional translations
	translations := []struct {
		tag         string
		translation string
	}{
		{
			tag:         "dir",
			translation: "{0} must point to an existing directory, but found '{1}'",
		},
		{
			tag:         "file",
			translation: "{0} must point to an existing file, but found '{1}'",
		},
		{
			tag:         "description",
			translation: fmt.Sprintf("must be less than %d", maxDescriptionLength),
		},
		{
			tag:         "name",
			translation: "is not a invalid name",
		},
	}
	for _, t := range translations {
		err = result.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation), translateFunc)
		if err != nil {
			panic(err)
		}
	}

	return &Validator{
		val:   result,
		data:  data,
		trans: trans,
	}
}

func registrationFunc(tag string, translation string) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) (err error) {
		if err = ut.Add(tag, translation, true); err != nil {
			return
		}

		return
	}
}

func translateFunc(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field(), reflect.ValueOf(fe.Value()).String())
	if err != nil {
		return fe.(error).Error()
	}

	return t
}

// validateDir checks if a given string is an existing directory.
func validateDir(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}

	return false
}

// validateFile checks if a given string is an existing file.
func validateFile(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return true
	}

	return false
}

// validateDescription checks if a given description is illegal.
func validateDescription(fl validator.FieldLevel) bool {
	description := fl.Field().String()

	return len(description) <= maxDescriptionLength
}

// validateName checks if a given name is illegal.
func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if errs := IsQualifiedName(name); len(errs) > 0 {
		return false
	}

	return true
}
