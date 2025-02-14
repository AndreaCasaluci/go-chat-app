package utils

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"regexp"
)

var Validator *validator.Validate = nil

func getValidator() *validator.Validate {
	if Validator == nil {
		Validator = validator.New()
		err :=Validator.RegisterValidation("usernamechars", validateUsernameChars)
		if err != nil {
			panic(err)
		}
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

func validateUsernameChars(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return re.MatchString(username)
}
