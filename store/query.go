// Copyright (c) 2016 SafetyCulture Pty Ltd. All Rights Reserved.

package store

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Stats struct {
	ValidNumbersCount     int `json:"valid_numbers_count"`
	FixedNumbersCount     int `json:"fixed_numbers_count"`
	InvalidNumbersCount   int `json:"invalid_numbers_count"`
	TotalNumbersProcessed int `json:"total_numbers_processed"`
}

type FileResults struct {
	ValidNumbers    []string      `json:"valid_numbers"`
	FixedNumbers    []FixedNumber `json:"fixed_numbers"`
	RejectedNumbers []string      `json:"rejected_numbers"`
}

// GetFileResults query DB for results from previously processed file
func (s *Store) GetFileResults(ref uuid.UUID) (*FileResults, error) {
	query := `SELECT number FROM numbers WHERE file_ref=$1`
	result := &FileResults{}
	rows, err := s.DB.Query(query, ref)
	if err != nil {
		log.Info(err)
		return nil, err
	}
	// iterate over each row
	for rows.Next() {
		var number string
		err = rows.Scan(&number)
		result.ValidNumbers = append(result.ValidNumbers, number)
	}

	query = `SELECT number FROM rejected_numbers WHERE file_ref=$1`
	rows, err = s.DB.Query(query, ref)
	if err != nil {
		log.Info(err)
		return nil, err
	}
	// iterate over each row
	for rows.Next() {
		var number string
		err = rows.Scan(&number)
		result.RejectedNumbers = append(result.RejectedNumbers, number)
	}

	query = `SELECT original_number, changes, fixed_number FROM fixed_numbers WHERE file_ref=$1`
	err = s.DB.Select(&result.FixedNumbers, query, ref)
	fmt.Printf("fixed numbs: %+v", result.FixedNumbers)
	return result, nil
}

// GetFileStats query DB for statistics from previously processed file
func (s *Store) GetFileStats(ref uuid.UUID) (*Stats, error) {
	query := `SELECT FROM numbers WHERE file_ref=$1;`
	result, err := s.DB.Exec(query, ref)
	if err != nil {
		return nil, err
	}
	validNumbers, err := result.RowsAffected()

	query = `SELECT FROM fixed_numbers WHERE file_ref=$1;`
	result, err = s.DB.Exec(query, ref)
	if err != nil {
		return nil, err
	}
	fixedNumbers, err := result.RowsAffected()

	query = `SELECT FROM rejected_numbers WHERE file_ref=$1;`
	result, err = s.DB.Exec(query, ref)
	if err != nil {
		return nil, err
	}
	rejectedNumbers, err := result.RowsAffected()

	return &Stats{
		ValidNumbersCount:     int(validNumbers),
		FixedNumbersCount:     int(fixedNumbers),
		InvalidNumbersCount:   int(rejectedNumbers),
		TotalNumbersProcessed: int(validNumbers) + int(fixedNumbers) + int(rejectedNumbers),
	}, nil
}

// SaveNumbers stores valid numbers
func (s *Store) SaveNumbers(numbers []Number) error {
	log.Infof("Saving %d numbers", len(numbers))
	log.Debugf("Numbers: %+v", numbers)
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("numbers", "number", "country_ioc_code", "file_ref"))
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveNumbers] unable to prepare pq.CopyIn"))
	}

	for _, num := range numbers {
		_, err = stmt.Exec(num.Number, num.CountryIOCCode, num.FileRef)
		if err != nil {
			log.Error(errors.Wrapf(err, "[SaveNumbers] unable to save number %+v", num))
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveNumbers] unable to execute bulk insert to numbers"))
	}

	err = stmt.Close()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveNumbers] unable to close DB connection"))
	}

	err = txn.Commit()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveNumbers] unable to commit bulk transaction"))
	}
	return nil
}

// SaveFixedNumbers stores fixed mobile numbers, the originally provided number, and a list of changes
func (s *Store) SaveFixedNumbers(fixedNums []FixedNumber) error {
	log.Infof("Saving %d fixed numbers", len(fixedNums))
	log.Debugf("Fixed Numbers: %+v", fixedNums)
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("fixed_numbers", "original_number", "changes", "fixed_number", "file_ref"))
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveFixedNumbers] unable to prepare pq.CopyIn"))
	}

	for _, num := range fixedNums {
		_, err = stmt.Exec(num.OriginalNumber, num.Changes, num.FixedNumber, num.FileRef)
		if err != nil {
			log.Error(errors.Wrapf(err, "[SaveFixedNumbers] unable to save number %+v", num))
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveFixedNumbers] unable to execute bulk insert to numbers"))
	}

	err = stmt.Close()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveFixedNumbers] unable to close DB connection"))
	}

	err = txn.Commit()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveFixedNumbers] unable to commit bulk transaction"))
	}
	return nil
}

// SaveRejectedNumbers saves invalid numbers that could not be fixed
func (s *Store) SaveRejectedNumbers(rejectedNums []RejectedNumber) error {
	log.Infof("Saving %d rejected numbers", len(rejectedNums))
	log.Infof("Numbers: %+v", rejectedNums)
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("rejected_numbers", "number", "file_ref"))
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveRejectedNumbers] unable to prepare pq.CopyIn"))
	}

	for _, num := range rejectedNums {
		_, err = stmt.Exec(num.Number, num.FileRef)
		if err != nil {
			log.Error(errors.Wrapf(err, "[SaveRejectedNumbers] unable to save number %+v", num))
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveRejectedNumbers] unable to execute bulk insert to numbers"))
	}

	err = stmt.Close()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveRejectedNumbers] unable to close DB connection"))
	}

	err = txn.Commit()
	if err != nil {
		log.Error(errors.Wrap(err, "[SaveRejectedNumbers] unable to commit bulk transaction"))
	}
	return nil
}
