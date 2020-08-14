package appconfig

import "time"

// Store settings
const (
	PostgresStore = "Postgres"
	MongoStore    = "Mongo"
)

// Web server settings
const (
	Port                   = 3000
	Store                  = MongoStore
	IdentityJSONSchemaPath = "file:////Users/trapck/go/krapi/model/schema.json"
)

// Postgres settings
const (
	PostgresDriver  = "postgres"
	PostgresConnStr = "user=postgres password=postgres dbname=postgres sslmode=disable"
)

// Mongo settings
const (
	MongoHost       = "mongodb://localhost:27017"
	MongoDBName     = "krapi"
	MongoDefTimeout = 5 * time.Second
)
