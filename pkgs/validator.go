package pkgs

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RequestValidator provides a reusable request validation logic.
type RequestValidator struct {
	validate *validator.Validate
}

// NewRequestValidator creates a new RequestValidator.
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		validate: validator.New(),
	}
}

// Validate validates the given request struct.
func (v *RequestValidator) Validate(c *gin.Context, req interface{}) error {
	if err := v.validate.Struct(req); err != nil {
		// Get the first validation error
		if validationErrors, ok := err.(validator.ValidationErrors); ok && len(validationErrors) > 0 {
			// Try to get custom error message
			field, _ := reflect.TypeOf(req).Elem().FieldByName(validationErrors[0].Field())
			message := field.Tag.Get("message")
			if message == "" {
				// If there is no custom message, use the default error
				Error(c, http.StatusBadRequest, validationErrors[0].Error())
			} else {
				// Use custom error message
				Error(c, http.StatusBadRequest, message)
			}
		} else {
			Error(c, http.StatusBadRequest, err.Error())
		}
		return err
	}
	return nil
}
