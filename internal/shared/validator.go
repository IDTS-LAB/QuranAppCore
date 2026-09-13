package shared

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(value any) error {
	return validate.Struct(value)
}

// ValidateWithDetails validates a struct and returns one friendly sentence
// per failing field, keyed by JSON field name, e.g.
// {"email": "Email must be a valid email address."}.
func ValidateWithDetails(value any) map[string]string {
	err := validate.Struct(value)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}

	names := jsonFieldNames(value)
	details := make(map[string]string, len(validationErrors))
	for _, ve := range validationErrors {
		key := ve.Field()
		display := prettyFieldName(ve.StructField())
		if name, ok := names[ve.StructField()]; ok {
			key = name
			display = prettyFieldName(name)
		}
		details[key] = friendlyValidationMessage(display, ve)
	}
	return details
}

// jsonFieldNames maps Go struct field names to their JSON names so error
// details reference the names clients actually send.
func jsonFieldNames(value any) map[string]string {
	names := map[string]string{}
	t := reflect.TypeOf(value)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return names
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := strings.Split(field.Tag.Get("json"), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		names[field.Name] = tag
	}
	return names
}

// prettyFieldName turns "display_name" or "DisplayName" into "Display name".
func prettyFieldName(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	words := strings.Fields(name)
	for i, word := range words {
		if i == 0 {
			words[i] = capitalize(word)
			continue
		}
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, " ")
}

func capitalize(word string) string {
	if word == "" {
		return word
	}
	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}

func friendlyValidationMessage(display string, ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return fmt.Sprintf("%s is required.", display)
	case "email":
		return fmt.Sprintf("%s must be a valid email address.", display)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long.", display, ve.Param())
	case "max":
		return fmt.Sprintf("%s must be no longer than %s characters.", display, ve.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long.", display, ve.Param())
	default:
		return fmt.Sprintf("%s is invalid.", display)
	}
}

func IsValidationError(err error) bool {
	var validationErrors validator.ValidationErrors

	return errors.As(err, &validationErrors)
}
