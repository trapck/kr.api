package main

import (
	"log"

	"github.com/trapck/kr.api/store"

	"github.com/trapck/kr.api/server"
)

func main() {
	s := store.PostgresStore{}
	if err := s.Init(); err != nil {
		log.Fatalf("could not open db connection %q", err)
	}
	defer s.Close()
	if err := server.NewApp(&s).Start(3000); err != nil {
		log.Fatalf("could not listen on port 3000 %v", err)
	}
}
