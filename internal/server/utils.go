package server

import (
	"encoding/csv"
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
)

// generates random UUID
// improvement: gererate unique hash against sorted CSV data. This can be used to avoid processing same file twice.
func generateHash(data [][]string) (uuid.UUID, error) {
	return uuid.NewV4()
}

// transform CSV into string slices
func readCSVFromHttpRequest(req *http.Request) ([][]string, error) {
	r := csv.NewReader(req.Body)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func handleError(w http.ResponseWriter, err error, code int) {
	log.Error(err)
	errJSON := jsonError{Msg: err.Error()}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errJSON)
}
