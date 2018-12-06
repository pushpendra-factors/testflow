package model

import (
	"errors"
	U "factors/util"
	"strings"
)

// IsValidName - Generic validator for names across entities.
func IsValidName(name string) error {
	// $ Prefix not allowed.
	if name == "" || strings.HasPrefix(name, U.NAME_PREFIX) {
		return errors.New("invalid name")
	}
	return nil
}
