package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMobileNumber(t *testing.T) {
	tests := map[string]struct {
		number   string
		expected *mobileNumber
		code     string
		err      error
	}{
		"valid number": {
			number: "27717278645",
			code:   "rsa",
			expected: &mobileNumber{
				NumberProvided:      "27717278645",
				FixedNumber:         "27717278645",
				countryAbbreviation: "rsa",
				Valid:               true,
				Changes:             "",
			},
			err: nil,
		},
		"fixable number by shortening": {
			number: "277172786457",
			code:   "rsa",
			expected: &mobileNumber{
				NumberProvided:      "277172786457",
				FixedNumber:         "27717278645",
				countryAbbreviation: "rsa",
				Valid:               false,
				Changes:             "shortened number by removing 7,",
			},
			err: nil,
		},
		"fixable number by prepending dialing code": {
			number: "717278645",
			code:   "rsa",
			expected: &mobileNumber{
				NumberProvided:      "717278645",
				FixedNumber:         "27717278645",
				countryAbbreviation: "rsa",
				Valid:               false,
				Changes:             "prepended number with 27,",
			},
			err: nil,
		},
		"invalid number because it is too short": {
			number: "271",
			code:   "rsa",
			expected: &mobileNumber{
				NumberProvided:      "271",
				FixedNumber:         "",
				countryAbbreviation: "rsa",
				Valid:               false,
				Changes:             "",
			},
			err: fmt.Errorf("invalid length 3, the length must be exactly 11"),
		},
		"invalid number because country code does not exist in configuration file": {
			number: "27717278645",
			code:   "DNE",
			expected: &mobileNumber{
				NumberProvided:      "27717278645",
				FixedNumber:         "",
				countryAbbreviation: "DNE",
				Valid:               false,
				Changes:             "",
			},
			err: fmt.Errorf("country IOC code DNE not found in lookup"),
		},
	}
	for tName, test := range tests {
		actual, err := newMobileNumber(test.code, test.number)
		if test.err == nil {
			require.NoError(t, err, tName)
		} else {
			require.Error(t, err, test.err.Error())
		}
		require.Equal(t, test.expected, actual, tName)
	}
}
