package main

import (
	"log"

	"github.com/nico-phil/go-log/internal/server"
)

func main() {
	s := server.New(8080)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start sever")
	}

}
