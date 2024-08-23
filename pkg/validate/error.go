package validate

import (
	"fmt"
	"gin-splitwise/pkg/logging"
	"reflect"

	"github.com/go-playground/validator/v10"
)

type ValidationErrDetail struct {
	Field string `json:"field"`
	Value interface{} `json:"value"`
	Message string `json:"message"`
}

type Test struct {
	Name string `json:"name"`
	Lastname string `json:"lastname"`
}

func ValidationErrorDetails(obj interface{}, tag string, errs validator.ValidationErrors) []*ValidationErrDetail {
	if len(errs) == 0 {
		return []*ValidationErrDetail{}
	}
	var errors []*ValidationErrDetail
	e := reflect.TypeOf(obj).Elem()
	for _, err := range errs {
		f, _ := e.FieldByName(err.Field())
		fieldName, _ := f.Tag.Lookup(tag)
		val := err.Value()
		var message string

		switch err.ActualTag() {
		case "required":
			message = fmt.Sprintf("required %s", fieldName)
		case "email":
			message = fmt.Sprintf("%s is not a valid email", fieldName)
		case "min":
			message = fmt.Sprintf("%s must be at least %s length", fieldName, err.Param())
		case "max":
			message = fmt.Sprintf("%s must be at most %s length", fieldName, err.Param())
		case "hexadecimal":
			message = fmt.Sprintf("%s must be a hexadecimal number", fieldName)
		case "gte":
			message = fmt.Sprintf("%s must be greater than or equal to %s", fieldName, err.Param())
		case "lte":
			message = fmt.Sprintf("%s must be less than or equal to %s", fieldName, err.Param())
		case "numeric":
			message = fmt.Sprintf("%s must be numeric", fieldName)
		default:
			logging.DefaultLogger().Warnf("unknown validation tag. tag:%s", err.ActualTag())
			message = fmt.Sprintf("invalid %s", fieldName)
		}
		errors = append(errors, &ValidationErrDetail{
			Field:   fieldName,
			Value:   val,
			Message: message,
		})
	}
	return errors
}

func NewValidationErrorDetails(field, message string, value interface{}) []*ValidationErrDetail {
	return []*ValidationErrDetail{
		{	
			Field: field,
			Value: value,
			Message: message,
		},
	}
}