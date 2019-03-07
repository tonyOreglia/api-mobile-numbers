package store

import (
	"database/sql"

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
		return nil, err
	}
	for rows.Next() {
		var number string
		err = rows.Scan(&number)
		result.ValidNumbers = append(result.ValidNumbers, number)
	}
	query = `SELECT number FROM rejected_numbers WHERE file_ref=$1`
	rows, err = s.DB.Query(query, ref)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var number string
		err = rows.Scan(&number)
		if err != nil {
			return nil, err
		}
		result.RejectedNumbers = append(result.RejectedNumbers, number)
	}
	query = `SELECT original_number, changes, fixed_number FROM fixed_numbers WHERE file_ref=$1`
	err = s.DB.Select(&result.FixedNumbers, query, ref)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetFileStats query DB for statistics from previously processed file
func (s *Store) GetFileStats(ref uuid.UUID) (*Stats, error) {
	query := `SELECT FROM numbers WHERE file_ref=$1`
	result, err := s.DB.Exec(query, ref)
	if err != nil {
		return nil, err
	}
	validNumbers, err := result.RowsAffected()

	query = `SELECT FROM fixed_numbers WHERE file_ref=$1`
	result, err = s.DB.Exec(query, ref)
	if err != nil {
		return nil, err
	}
	fixedNumbers, err := result.RowsAffected()

	query = `SELECT FROM rejected_numbers WHERE file_ref=$1`
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
	if len(numbers) == 0 {
		return nil
	}
	log.Infof("Saving %d numbers", len(numbers))
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("numbers", "number", "country_ioc_code", "file_ref"))
	if err != nil {
		return errors.Wrap(err, "[SaveNumbers] unable to prepare pq.CopyIn")
	}

	for _, num := range numbers {
		_, err = stmt.Exec(num.Number, num.CountryIOCCode, num.FileRef)
		if err != nil {
			endTrasaction(stmt, txn)
			return errors.Wrapf(err, "[SaveNumbers] unable to save number %+v", num)
		}
	}
	return executeTransaction(stmt, txn, "SaveNumbers")
}

// SaveFixedNumbers stores fixed mobile numbers, the originally provided number, and a list of changes
func (s *Store) SaveFixedNumbers(fixedNums []FixedNumber) error {
	if len(fixedNums) == 0 {
		return nil
	}
	log.Infof("Saving %d fixed numbers", len(fixedNums))
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("fixed_numbers", "original_number", "changes", "fixed_number", "file_ref"))
	if err != nil {
		endTrasaction(stmt, txn)
		return errors.Wrap(err, "[SaveFixedNumbers] unable to prepare pq.CopyIn")
	}
	for _, num := range fixedNums {
		_, err = stmt.Exec(num.OriginalNumber, num.Changes, num.FixedNumber, num.FileRef)
		if err != nil {
			endTrasaction(stmt, txn)
			return errors.Wrapf(err, "[SaveFixedNumbers] unable to save number %+v", num)
		}
	}
	return executeTransaction(stmt, txn, "SaveFixedNumbers")
}

// SaveRejectedNumbers saves invalid numbers that could not be fixed
func (s *Store) SaveRejectedNumbers(rejectedNums []RejectedNumber) error {
	if len(rejectedNums) == 0 {
		return nil
	}
	log.Infof("Saving %d rejected numbers", len(rejectedNums))
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("rejected_numbers", "number", "file_ref"))
	if err != nil {
		endTrasaction(stmt, txn)
		return errors.Wrap(err, "[SaveRejectedNumbers] unable to prepare pq.CopyIn")
	}
	for _, num := range rejectedNums {
		_, err = stmt.Exec(num.Number, num.FileRef)
		if err != nil {
			endTrasaction(stmt, txn)
			return errors.Wrapf(err, "[SaveRejectedNumbers] unable to save number %+v", num)
		}
	}
	return executeTransaction(stmt, txn, "SaveRejectedNumbers")
}

func executeTransaction(stmt *sql.Stmt, txn *sql.Tx, op string) error {
	_, err := stmt.Exec()
	if err != nil {
		endTrasaction(stmt, txn)
		return errors.Wrapf(err, "[%s] unable to execute bulk insert to numbers", op)
	}

	err = stmt.Close()
	if err != nil {
		if e := stmt.Close(); e != nil {
			log.Error(e)
		}
		return errors.Wrapf(err, "[%s] unable to close DB connection", op)
	}

	err = txn.Commit()
	if err != nil {
		return errors.Wrapf(err, "[%s] unable to commit bulk transaction", op)
	}
	return nil
}

func endTrasaction(stmt *sql.Stmt, txn *sql.Tx) {
	if e := stmt.Close(); e != nil {
		log.Error(e)
	}
	if e := txn.Commit(); e != nil {
		log.Error(e)
	}
}
