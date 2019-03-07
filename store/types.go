package store

import (
	"github.com/gofrs/uuid"
)

// Number is used in query to store valid numer in DB
type Number struct {
	Number         string    `db:"number"`
	CountryIOCCode string    `db:"country_ioc_code"`
	FileRef        uuid.UUID `db:"file_ref"`
}

// FixedNumber is used in query to store fixed number in DB
type FixedNumber struct {
	OriginalNumber string    `json:"original_number" db:"original_number"`
	Changes        string    `json:"changes" db:"changes"`
	FixedNumber    string    `json:"fixed_number" db:"fixed_number"`
	FileRef        uuid.UUID `json:"-" db:"file_ref"`
}

// RejectedNumber is used in query to store rejected number in DB
type RejectedNumber struct {
	Number  string    `db:"number"`
	FileRef uuid.UUID `db:"file_ref"`
}
