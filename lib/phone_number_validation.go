package lib

import (
	"fmt"
	"regexp"
)

func ValidatePhoneNumber(phone string) error {
	pattern := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !pattern.MatchString(phone) {
		return fmt.Errorf("invalid phone number")
	}
	return nil
}
