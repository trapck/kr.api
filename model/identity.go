package model

import (
	"time"
)

//Address describes a base address model
type Address struct {
	ID    string `db:"id"`
	Value string `db:"value"`
	Via   string `db:"via"`
}

//RecoveryAddress is a recovery address model
type RecoveryAddress struct {
	Address
	Identity string `db:"identity"`
}

//VerifiableAddress is a verifiable address model
type VerifiableAddress struct {
	Address
	ExpiresAt  *time.Time `db:"expires_at"`
	Verified   bool       `db:"verified"`
	VerifiedAt *time.Time `db:"verified_at"`
	Identity   string     `db:"identity"`
}

//Identity is an indentity model
type Identity struct {
	ID                  string `db:"id"`
	RecoveryAddresses   []RecoveryAddress
	SchemaID            string `db:"schema_id"`
	SchemaURL           string `db:"schema_url"`
	VerifiableAddresses []VerifiableAddress
}
