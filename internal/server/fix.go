package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	log "github.com/sirupsen/logrus"
)

type requirements struct {
	countryCode string
	length      int
}

// requirements by country configuration
var lookupRequirements = map[string]requirements{
	"rsa": {countryCode: "27", length: 11},
	"aus": {countryCode: "61", length: 9},
	"por": {countryCode: "351", length: 12},
	"usa": {countryCode: "1", length: 11},
}

// fix attempts to fix a given mobile number to adhere to the requirments for a given country
// if it cannot fix the number, an error is returned
func (n *mobileNumber) fix() error {
	// if country IOC code is not found in requirements lookup, this number is rejected
	req, found := lookupRequirements[n.countryAbbreviation]
	if !found {
		n.Valid = false
		n.FixedNumber = ""
		return &jsonError{Msg: fmt.Sprintf("country IOC code %s not found in lookup", n.countryAbbreviation)}
	}

	if !n.dialingCodeIsCorrect(req.countryCode) {
		n.Valid = false
		n.prependDialingCodeFix(req.countryCode)
	}

	if !n.onlyDigitsInNumber() {
		n.Valid = false
		n.removeNonDigitsFix()
	}

	if n.numberIsTooLong(req.length) {
		n.Valid = false
		n.shortenNumberFix(req.length)
	}

	// This number is rejected if to short
	if n.numberIsTooShort(req.length) {
		n.Valid = false
		n.Changes = ""
		n.FixedNumber = ""
		errString := fmt.Sprintf("invalid length %d, the length must be exactly %d", len(n.NumberProvided), req.length)
		log.Error(&jsonError{Msg: errString})
		return &jsonError{Msg: errString}
	}
	return nil
}

func (n *mobileNumber) dialingCodeIsCorrect(code string) bool {
	return strings.HasPrefix(n.FixedNumber, code)
}

func (n *mobileNumber) prependDialingCodeFix(code string) {
	n.FixedNumber = fmt.Sprintf("%s%s", code, n.FixedNumber)
	n.Changes = fmt.Sprintf("%s%s,", n.Changes, fmt.Sprintf("prepended number with %s", code))
}

func (n *mobileNumber) onlyDigitsInNumber() bool {
	err := validation.ValidateStruct(n, validation.Field(&n.FixedNumber, validation.Required, is.Digit))
	if err != nil {
		return false
	}
	return true
}

func (n *mobileNumber) removeNonDigitsFix() {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Error("unable to gererate regex")
	}
	n.FixedNumber = reg.ReplaceAllString(n.FixedNumber, "")
	n.Changes = fmt.Sprintf("%s%s,", n.Changes, "removed non digits from number")
}

func (n *mobileNumber) numberIsTooLong(requiredLength int) bool {
	return len(n.FixedNumber) > requiredLength
}

func (n *mobileNumber) shortenNumberFix(requiredLength int) {
	digitsToRemove := n.FixedNumber[requiredLength:len(n.FixedNumber)]
	n.FixedNumber = n.FixedNumber[0:requiredLength]
	changeString := fmt.Sprintf("shortened number by removing %s", digitsToRemove)
	n.Changes = fmt.Sprintf("%s%s,", n.Changes, changeString)
	log.Info(changeString)
}

func (n *mobileNumber) numberIsTooShort(requiredLength int) bool {
	return len(n.FixedNumber) < requiredLength
}
