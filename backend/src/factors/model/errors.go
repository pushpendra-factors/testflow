package model

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var InvalidProjectScopeError = errors.New("Invalid project scope.")

// IsInvalidProjectError returns current error has invalid project scope error.
func IsInvalidProjectError(err error) bool {
	if errs, ok := err.(gorm.Errors); ok {
		for _, err := range errs {
			if err == InvalidProjectScopeError {
				return true
			}
		}
	}
	return err == InvalidProjectScopeError
}
