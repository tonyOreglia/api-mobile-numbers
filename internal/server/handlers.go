package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/tonyOreglia/api-mobile-numbers/store"
)

type fileData struct {
	Ref   uuid.UUID   `json:"ref"`
	Stats store.Stats `json:"stats"`
	Href  string      `json:"href"`
}

// Query server to generate statistical information about a previously processed file
func (s *Server) getFileDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ref := vars["ref"]
	refUUID, err := uuid.FromString(ref)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	stats, err := s.db.GetFileStats(refUUID)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(fileData{
		Ref:   refUUID,
		Stats: *stats,
		Href:  buildHref(url, port, ref),
	})
}

// return downloadable data from previously processed file
func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ref := vars["ref"]
	refUUID, err := uuid.FromString(ref)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	res, err := s.db.GetFileResults(refUUID)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("Attachment; filename=%s.json", ref))
	json.NewEncoder(w).Encode(res)
}

// test validity of a single mobile number
func testNumberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	num, err := newMobileNumber(vars["countryAbbreviation"], vars["number"])
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if num.Valid {
		json.NewEncoder(w).Encode(struct{ Valid bool }{true})
		return
	}
	json.NewEncoder(w).Encode(num)
}

// process a CSV payload of mobile numbers
func (s *Server) storeNumbersHandler(w http.ResponseWriter, r *http.Request) {
	var (
		numbers         []store.Number
		fixedNumbers    []store.FixedNumber
		rejectedNumbers []store.RejectedNumber
	)
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	csvPayload, err := readCSVFromHttpRequest(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}
	hash, err := generateHash(csvPayload)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
	fmt.Println(hash)
	for i, row := range csvPayload {
		if i == 0 {
			continue // skip first row
		}
		num, err := newMobileNumber(vars["countryAbbreviation"], row[1])
		if err != nil {
			rejectedNumber := store.RejectedNumber{
				Number:  num.NumberProvided,
				FileRef: hash,
			}
			rejectedNumbers = append(rejectedNumbers, rejectedNumber)
			continue
		}
		if num.Valid {
			number := store.Number{
				Number:         num.NumberProvided,
				FileRef:        hash,
				CountryIOCCode: vars["countryAbbreviation"],
			}
			numbers = append(numbers, number)
			continue
		}
		fixedNumber := store.FixedNumber{
			OriginalNumber: num.NumberProvided,
			FixedNumber:    num.FixedNumber,
			Changes:        strings.Join(num.Changes, (", ")),
			FileRef:        hash,
		}
		fixedNumbers = append(fixedNumbers, fixedNumber)
	}

	err = s.db.SaveNumbers(numbers)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	err = s.db.SaveFixedNumbers(fixedNumbers)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	err = s.db.SaveRejectedNumbers(rejectedNumbers)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	resp := fileData{
		Ref: hash,
		Stats: store.Stats{
			ValidNumbersCount:     len(numbers),
			FixedNumbersCount:     len(fixedNumbers),
			InvalidNumbersCount:   len(rejectedNumbers),
			TotalNumbersProcessed: len(numbers) + len(fixedNumbers) + len(rejectedNumbers),
		},
		Href: buildHref(url, port, hash.String()),
	}
	json.NewEncoder(w).Encode(resp)
}

// helpfer function to build URL that user can use in future call
// to download results of a processed file
func buildHref(url string, port int, fileRef string) string {
	return fmt.Sprintf("%s:%d/numbers/%s", url, port, fileRef)
}
