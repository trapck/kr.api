package store

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // db driver
	"github.com/trapck/kr.api/model"
)

//PostgresStore is postgres storage implementation
type PostgresStore struct {
	db *sqlx.DB
}

// Init initializes connetion
func (s *PostgresStore) Init() error {
	db, e := sqlx.Connect("postgres", "user=postgres password=postgres dbname=postgres sslmode=disable")
	s.db = db
	return e
}

// Close closes connetion
func (s *PostgresStore) Close() error {
	if success, e := s.ensureConnection(); !success {
		return e
	}
	return nil
}

// List returns all identities
func (s *PostgresStore) List() ([]model.Identity, error) {
	verifiableAddresses := []model.VerifiableAddress{}
	recoveryAddresses := []model.RecoveryAddress{}
	identities := []model.Identity{}
	e := s.db.Select(&identities, "SELECT * FROM identity")
	if e != nil {
		return nil, e
	}
	e = s.db.Select(&verifiableAddresses, "SELECT * FROM verifiable_address")
	if e != nil {
		return nil, e
	}
	e = s.db.Select(&recoveryAddresses, "SELECT * FROM recovery_address")
	if e != nil {
		return nil, e
	}
	return s.mapIdentityEntities(identities, verifiableAddresses, recoveryAddresses), nil
}

// Get returns all identities
func (s *PostgresStore) Get(id string) (model.Identity, error) {
	identitiy := model.Identity{}
	e := s.db.Get(&identitiy, "SELECT * FROM identity WHERE id = $1", id)
	if e != nil {
		return identitiy, e
	}
	va := []model.VerifiableAddress{}
	e = s.db.Select(&va, "SELECT * FROM verifiable_address WHERE identity=$1", id)
	if e != nil {
		return identitiy, e
	}
	identitiy.VerifiableAddresses = va
	ra := []model.RecoveryAddress{}
	e = s.db.Select(&ra, "SELECT * FROM recovery_address WHERE identity=$1", id)
	if e != nil {
		return identitiy, e
	}
	identitiy.RecoveryAddresses = ra
	return identitiy, e
}

// Create inserts identity
func (s *PostgresStore) Create(i model.Identity) (model.Identity, error) {
	t, e := s.db.Begin()
	if e != nil {
		return i, e
	}
	e = s.insertIdentity(t, i)
	if e != nil {
		t.Rollback()
		return i, e
	}
	if len(i.RecoveryAddresses) > 0 {
		e = s.insertRecoveryAddresses(t, i.ID, i.RecoveryAddresses)
		if e != nil {
			t.Rollback()
			return i, e
		}
	}

	if len(i.VerifiableAddresses) > 0 {
		e = s.insertVerifiableAddresses(t, i.ID, i.VerifiableAddresses)
		if e != nil {
			t.Rollback()
			return i, e
		}
	}
	t.Commit()
	return i, nil
}

// Update updates identity
func (s *PostgresStore) Update(id string, i model.Identity) (model.Identity, error) {
	e := s.Delete(id)
	if e != nil {
		return i, e
	}
	//TODO: wrap delete and insert into single transaction
	return s.Create(i)
}

//Delete deletes identity
func (s *PostgresStore) Delete(id string) error {
	t, e := s.db.Begin()
	if e != nil {
		return e
	}
	_, e = t.Exec("DELETE FROM verifiable_address WHERE identity = $1", id)
	if e != nil {
		t.Rollback()
		return e
	}
	_, e = t.Exec("DELETE FROM recovery_address WHERE identity = $1", id)
	if e != nil {
		t.Rollback()
		return e
	}
	_, e = t.Exec("DELETE FROM identity WHERE id = $1", id)
	if e != nil {
		t.Rollback()
		return e
	}
	t.Commit()
	return nil
}

func (s *PostgresStore) ensureConnection() (isConnected bool, e error) {
	isConnected = s.db != nil
	if !isConnected {
		e = fmt.Errorf("db connection is not initialized")
	}
	return
}

func (s *PostgresStore) insertIdentity(t *sql.Tx, i model.Identity) error {
	_, e := t.Exec("INSERT INTO identity (id, schema_id, schema_url) VALUES ($1, $2, $3)", i.ID, i.SchemaID, i.SchemaURL)
	return e
}

func (s *PostgresStore) insertRecoveryAddresses(t *sql.Tx, identity string, a []model.RecoveryAddress) error {
	cnt := len(a)
	q := "INSERT INTO recovery_address (id, value, via, identity) VALUES "
	p := []interface{}{}
	for index, a := range a {
		pos := index * 4
		q += fmt.Sprintf("($%d,$%d,$%d,$%d)", pos+1, pos+2, pos+3, pos+4)
		if index != cnt-1 {
			q += ","
		}
		p = append(p, a.ID, a.Value, a.Via, identity)
	}
	_, e := t.Exec(q, p...)
	return e
}

func (s *PostgresStore) insertVerifiableAddresses(t *sql.Tx, identity string, a []model.VerifiableAddress) error {
	cnt := len(a)
	q := "INSERT INTO verifiable_address (id, value, via, verified, verified_at, expires_at, identity) VALUES "
	p := []interface{}{}
	for index, a := range a {
		pos := index * 7
		q += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)", pos+1, pos+2, pos+3, pos+4, pos+5, pos+6, pos+7)
		if index != cnt-1 {
			q += ","
		}
		p = append(p, a.ID, a.Value, a.Via, a.Verified, a.VerifiedAt, a.ExpiresAt, identity)
	}
	_, e := t.Exec(q, p...)
	return e
}

func (s *PostgresStore) findIdentity(id string, identities []model.Identity) (model.Identity, error) {
	for _, i := range identities {
		if i.ID == id {
			return i, nil
		}
	}
	return model.Identity{}, fmt.Errorf("not found")
}

func (s *PostgresStore) mapIdentityEntities(identities []model.Identity, verifiableAddresses []model.VerifiableAddress, recoveryAddresses []model.RecoveryAddress) []model.Identity {
	iMap := map[string]model.Identity{}
	for _, a := range verifiableAddresses {
		_, ok := iMap[a.Identity]
		if !ok {
			i, _ := s.findIdentity(a.Identity, identities)
			iMap[a.Identity] = i
		}
		va := iMap[a.Identity].VerifiableAddresses
		va = append(va, a)
	}
	for _, a := range recoveryAddresses {
		_, ok := iMap[a.Identity]
		if !ok {
			i, _ := s.findIdentity(a.Identity, identities)
			iMap[a.Identity] = i
		}
		ra := iMap[a.Identity].RecoveryAddresses
		ra = append(ra, a)
	}

	result := []model.Identity{}
	for _, i := range identities {
		if mappedIdentity, ok := iMap[i.ID]; ok {
			result = append(result, mappedIdentity)
		} else {
			result = append(result, i)
		}
	}
	return result
}
