package model

import (
	"time"
)

//Address describes a base address model
type Address struct {
	ID    string `json:"id" db:"id" bson:"id"`
	Value string `json:"value" db:"value" bson:"value"`
	Via   string `json:"via" db:"via" bson:"via"`
}

//RecoveryAddress is a recovery address model
type RecoveryAddress struct {
	Address  `bson:",inline"`
	Identity string `json:"-" db:"identity" bson:",-,omitempty"`
}

//VerifiableAddress is a verifiable address model
type VerifiableAddress struct {
	Address    `bson:",inline"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at" bson:"expires_at,omitempty"`
	Verified   bool       `json:"verified" db:"verified" bson:"verified"`
	VerifiedAt *time.Time `json:"verified_at" db:"verified_at" bson:"verified_at,omitempty"`
	Identity   string     `json:"-" db:"identity" bson:",-,omitempty"`
}

//Identity is an indentity model
type Identity struct {
	ID                  string              `json:"id" db:"id" bson:"id"`
	RecoveryAddresses   []RecoveryAddress   `json:"recovery_addresses" bson:"recovery_addresses,omitempty"`
	SchemaID            string              `json:"schema_id" db:"schema_id" bson:"schema_id"`
	SchemaURL           string              `json:"schema_url" db:"schema_url" bson:"schema_url"`
	VerifiableAddresses []VerifiableAddress `json:"verifiable_addresses" bson:"verifiable_address,omitempty"`
}
