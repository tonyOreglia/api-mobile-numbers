package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type jsonError struct {
	Msg string `json:"message"`
}

func (e *jsonError) Error() string {
	return e.Msg
}

type requirements struct {
	countryCode string
	length      int
}

type validator struct {
	num     *mobileNumber
	iocCode string
	req     requirements
}

var lookupRequirements = map[string]requirements{
	"rsa": {countryCode: "27", length: 11},
	"aus": {countryCode: "61", length: 9},
	"por": {countryCode: "351", length: 12},
	"usa": {countryCode: "1", length: 11},
}

func validate(n *mobileNumber) (error, int) {
	req, found := lookupRequirements[n.iocCode]
	if !found {
		return &jsonError{Msg: fmt.Sprintf("country IOC code %s not found in lookup", n.iocCode)}, http.StatusNotFound
	}

	if !strings.HasPrefix(n.Number, req.countryCode) {
		return &jsonError{Msg: fmt.Sprintf("%s mobile number must start with %s", n.iocCode, req.countryCode)}, http.StatusBadRequest
	}

	err := validation.ValidateStruct(n,
		validation.Field(&n.Number,
			validation.Required,
			is.Digit,
			validation.Length(req.length, req.length).Error(fmt.Sprintf("invalid length %d, the length must be exactly 11", len(n.Number))),
		),
	)
	if err != nil {
		return err, http.StatusBadRequest
	}
	return nil, http.StatusOK

}
