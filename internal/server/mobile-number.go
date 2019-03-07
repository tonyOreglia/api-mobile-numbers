package server

// information regarding valid and fixed mobile numbers
type mobileNumber struct {
	NumberProvided      string `json:"number_provided"`
	FixedNumber         string `json:"number_fixed"`
	countryAbbreviation string
	Valid               bool     `json:"valid"`
	Changes             []string `json:"changes"`
}

// Generates a mobile number data object after validating and attempting to fix
// if the number could not be fixed, an error is returned
func newMobileNumber(countryAbbreviation string, number string) (*mobileNumber, error) {
	mobileNum := &mobileNumber{
		NumberProvided:      number,
		FixedNumber:         number,
		countryAbbreviation: countryAbbreviation,
		Valid:               true,
	}
	err := mobileNum.fix()
	return mobileNum, err
}
