package utils

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate = nil

func getValidator() *validator.Validate{
	if Validator == nil {
		Validator = validator.New()
	}
	return Validator
}

func ValidateStruct(data interface{}) error {
	validate := getValidator()

	err := validate.Struct(data)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				return fmt.Errorf("field %s failed validation: %s", fieldError.Field(), fieldError.Tag())
			}
		}
		return fmt.Errorf("validation failed: %v", err)
	}

	return nil
}