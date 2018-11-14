package model

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var ErrInvalidProjectScope = errors.New("invalid project scope")

// IsInvalidProjectScopeError returns current error has invalid project scope error.
func isInvalidProjectScopeError(err error) bool {
	if errs, ok := err.(gorm.Errors); ok {
		for _, err := range errs {
			if err == ErrInvalidProjectScope {
				return true
			}
		}
	}
	return err == ErrInvalidProjectScope
}

// IsValidProjectScope return false if projectId is invalid.
func isValidProjectScope(projectId uint64) bool {
	return projectId != 0
}
