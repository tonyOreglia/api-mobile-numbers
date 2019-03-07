package store

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func PrepareMockStore(t *testing.T) (*sqlx.DB, *Store, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	db := sqlx.NewDb(mockDB, "sqlmock")

	dbStore := &Store{
		DB: db,
	}
	return db, dbStore, mock
}

func TestGetFileResults(t *testing.T) {
	testUUID, err := uuid.NewV4()
	require.NoError(t, err)
	db, DBStore, mock := PrepareMockStore(t)
	defer db.Close()
	mock.ExpectQuery(`SELECT number FROM numbers WHERE file_ref=\$1`).
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"number", "file_ref", "country_ioc_code"}).AddRow("1234", testUUID, "rsa"))

	mock.ExpectQuery(`SELECT number FROM rejected_numbers WHERE file_ref=\$1`).
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"number"}).AddRow("1234"))

	mock.ExpectQuery(`SELECT original_number, changes, fixed_number FROM fixed_numbers WHERE file_ref=\$1`).
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"original_number", "changes", "fixed_number", "file_ref"}).AddRow("1234", "change1,chang2", "1234", testUUID))

	DBStore.GetFileResults(testUUID)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
