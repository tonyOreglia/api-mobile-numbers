package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tonyOreglia/api-mobile-numbers/store"
)

var (
	url  = "http://localhost"
	port = 80
)

// Server defines a HTTP Server
type Server struct {
	r  *mux.Router
	db *store.Store
}

// New returns HTTP Server configured for localhost port 80
func New() *Server {
	server := new(Server)
	server.db = store.New("postgresql://olx:olx@localhost/mobile_numbers?sslmode=disable", 2)
	server.r = mux.NewRouter()
	server.r.HandleFunc("/{countryAbbreviation}/numbers/test/{number}", testNumberHandler).
		Methods("POST")
	server.r.HandleFunc("/{countryAbbreviation}/numbers", server.storeNumbersHandler).
		Methods("POST")
	server.r.HandleFunc("/numbers/results/{ref}", server.getFileDetailsHandler)
	server.r.HandleFunc("/numbers/{ref}", server.downloadHandler)
	return server
}

// Start starts the server
func (s *Server) Start() error {
	return http.ListenAndServe(":80", s.r)
}
