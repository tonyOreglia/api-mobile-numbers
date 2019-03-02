package server

import (
	"net/http"
)

type mobileNumber struct {
	Number  string `json:"number"`
	iocCode string
}

func NewMobileNumber(countryIOCCode string, number string) (*mobileNumber, error, int) {
	mobileNumb := &mobileNumber{
		Number:  number,
		iocCode: countryIOCCode,
	}

	if err, code := validate(mobileNumb); err != nil {
		return nil, err, code
	}

	return mobileNumb, nil, http.StatusOK
}
