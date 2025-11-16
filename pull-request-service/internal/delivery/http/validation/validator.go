package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateStatusEnum(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	return status == "OPEN" || status == "MERGED"
}

type validatorImpl struct {
	validate *validator.Validate
}

func (v *validatorImpl) Validate(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		var errMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, formatError(err))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errMsgs, ", "))
	}
	return nil
}

func NewValidator() *validatorImpl {
	v := validator.New()
	if err := v.RegisterValidation("status_enum", validateStatusEnum); err != nil {
		panic(err)
	}
	return &validatorImpl{validate: v}
}

func formatError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "status_enum":
		return fmt.Sprintf("%s must be one of: OPEN, MERGED", field)
	case "max":
		if err.Kind().String() == "string" {
			return fmt.Sprintf("%s must be at most %s characters", field, err.Param())
		}
		return fmt.Sprintf("%s must have at most %s items", field, err.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
