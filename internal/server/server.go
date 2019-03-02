package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Server defines a HTTP Server
type Server struct {
	r *mux.Router
}

func New() *Server {
	server := new(Server)
	server.r = mux.NewRouter()
	server.r.HandleFunc("/{countryCode}/test/{number}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		number, err, code := NewMobileNumber(vars["countryCode"], vars["number"])
		if err != nil {
			log.Error(err)
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(err)
			return
		}
		json.NewEncoder(w).Encode(number)
	})
	return server
}

func (s Server) Start() error {
	return http.ListenAndServe(":80", s.r)
}
