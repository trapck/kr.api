package main

import (
	"log"

	"github.com/trapck/kr.api/appconfig"
	"github.com/trapck/kr.api/mongostore"
	"github.com/trapck/kr.api/postgresstore"
	"github.com/trapck/kr.api/server"
)

func main() {
	var s server.Store
	switch appconfig.Store {
	case appconfig.PostgresStore:
		ps := &postgresstore.Store{}
		if err := ps.Init(); err != nil {
			log.Fatalf("could not open postgres db connection %q", err)
		}
		s = ps
		defer ps.Close()
	case appconfig.MongoStore:
		ms := &mongostore.Store{}
		if err := ms.Init(); err != nil {
			log.Fatalf("could not open mongo db connection %q", err)
		}
		s = ms
		defer ms.Close()
	default:
		log.Fatalf("unknown store type")
	}
	if err := server.NewApp(s).Start(appconfig.Port); err != nil {
		log.Fatalf("could not listen on port %d %v", appconfig.Port, err)
	}
}
