package model

import (
	"time"
)

//Address describes a base address model
type Address struct {
	ID    string `json:"id" db:"id"`
	Value string `json:"value" db:"value"`
	Via   string `json:"via" db:"via"`
}

//RecoveryAddress is a recovery address model
type RecoveryAddress struct {
	Address
	Identity string `db:"identity"`
}

//VerifiableAddress is a verifiable address model
type VerifiableAddress struct {
	Address
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
	Verified   bool       `json:"verified" db:"verified"`
	VerifiedAt *time.Time `json:"verified_at" db:"verified_at"`
	Identity   string     `db:"identity"`
}

//Identity is an indentity model
type Identity struct {
	ID                  string              `json:"id" db:"id"`
	RecoveryAddresses   []RecoveryAddress   `json:"recovery_addresses"`
	SchemaID            string              `json:"schema_id" db:"schema_id"`
	SchemaURL           string              `json:"schema_url" db:"schema_url"`
	VerifiableAddresses []VerifiableAddress `json:"verifiable_addresses"`
}
