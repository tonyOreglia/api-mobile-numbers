package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/tonyOreglia/api-mobile-numbers/internal/server"
)

func main() {
	server := server.New()
	log.Fatal(server.Start())
}
