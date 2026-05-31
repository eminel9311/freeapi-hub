package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validate là instance singleton của validator.
// Init 1 lần lúc start app, reuse khắp nơi (thread-safe).
var Validate = validator.New()

// ValidateStruct validate struct với tags.
//
// Ví dụ sử dụng:
//
//	type RegisterReq struct {
//	    Email    string `json:"email" validate:"required,email"`
//	    Password string `json:"password" validate:"required,min=8"`
//	}
//	if err := validator.ValidateStruct(req); err != nil {
//	    return err
//	}
func ValidateStruct(s any) error {
	return Validate.Struct(s)
}
