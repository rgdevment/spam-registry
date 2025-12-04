package http

import (
	"errors"
	"strings"
)

type CreateReportRequest struct {
	PhoneNumber string `json:"phone_number"`
	Category    string `json:"category"`
	Comment     string `json:"comment"`
}

func (r *CreateReportRequest) Validate() error {
	if len(r.PhoneNumber) < 5 {
		return errors.New("phone_number is too short")
	}

	validCategories := map[string]bool{
		"SPAM": true, "FRAUD": true, "PHISHING": true,
		"DEBT_COLLECTION": true, "SALES": true,
	}

	if !validCategories[strings.ToUpper(r.Category)] {
		return errors.New("invalid category")
	}

	return nil
}
