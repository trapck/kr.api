package postgresstore

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // db driver
	"github.com/trapck/kr.api/appconfig"
	"github.com/trapck/kr.api/model"
)

//Store is postgres storage implementation
type Store struct {
	db *sqlx.DB
}

// Init initializes connetion
func (s *Store) Init() error {
	db, e := sqlx.Connect(appconfig.PostgresDriver, appconfig.PostgresConnStr)
	s.db = db
	return e
}

// Close closes connetion
func (s *Store) Close() error {
	if success, e := s.ensureConnection(); !success {
		return e
	}
	return nil
}

// List returns all identities
func (s *Store) List() ([]model.Identity, error) {
	verifiableAddresses := []model.VerifiableAddress{}
	recoveryAddresses := []model.RecoveryAddress{}
	identities := []model.Identity{}
	e := s.db.Select(&identities, "SELECT * FROM identity")
	if e != nil {
		return nil, e
	}
	e = s.db.Select(&verifiableAddresses, "SELECT * FROM verifiable_address")
	if e != nil && !s.NoRows(e) {
		return nil, e
	}
	e = s.db.Select(&recoveryAddresses, "SELECT * FROM recovery_address")
	if e != nil && !s.NoRows(e) {
		return nil, e
	}
	return s.mapIdentityEntities(identities, verifiableAddresses, recoveryAddresses), nil
}

// Get returns all identities
func (s *Store) Get(id string) (model.Identity, error) {
	identitiy := model.Identity{}
	e := s.db.Get(&identitiy, "SELECT * FROM identity WHERE id = $1", id)
	if e != nil {
		return identitiy, e
	}
	va := []model.VerifiableAddress{}
	e = s.db.Select(&va, "SELECT * FROM verifiable_address WHERE identity=$1", id)
	if e != nil && !s.NoRows(e) {
		return identitiy, e
	}
	identitiy.VerifiableAddresses = va
	ra := []model.RecoveryAddress{}
	e = s.db.Select(&ra, "SELECT * FROM recovery_address WHERE identity=$1", id)
	if e != nil && !s.NoRows(e) {
		return identitiy, e
	}
	identitiy.RecoveryAddresses = ra
	return identitiy, e
}

// Create inserts identity
func (s *Store) Create(i model.Identity) (model.Identity, error) {
	t, e := s.createTx(&i, nil)
	if e != nil && t != nil {
		t.Rollback()
		return i, e
	}

	return i, t.Commit()
}

func (s *Store) createTx(i *model.Identity, existingTx *sql.Tx) (*sql.Tx, error) {
	var e error
	if existingTx == nil {
		existingTx, e = s.db.Begin()
	}
	if e != nil {
		return existingTx, e
	}
	e = s.insertIdentity(existingTx, *i)
	if e != nil {
		return existingTx, e
	}
	if len(i.RecoveryAddresses) > 0 {
		e = s.insertRecoveryAddresses(existingTx, i.ID, i.RecoveryAddresses)
		if e != nil {
			return existingTx, e
		}
	}

	if len(i.VerifiableAddresses) > 0 {
		e = s.insertVerifiableAddresses(existingTx, i.ID, i.VerifiableAddresses)
		if e != nil {
			return existingTx, e
		}
	}
	return existingTx, nil
}

// Update updates identity
func (s *Store) Update(id string, i model.Identity) (model.Identity, error) {
	return i, s.execTxChain(
		func(t *sql.Tx) error {
			_, e := s.deleteTx(id, t)
			return e
		}, func(t *sql.Tx) error {
			_, e := s.createTx(&i, t)
			return e
		},
	)
}

//Delete deletes identity
func (s *Store) Delete(id string) error {
	t, e := s.deleteTx(id, nil)
	if e != nil && t != nil {
		t.Rollback()
		return e
	}
	return t.Commit()
}

//Delete deletes identity
func (s *Store) deleteTx(id string, existingTx *sql.Tx) (*sql.Tx, error) {
	var e error
	if existingTx == nil {
		existingTx, e = s.db.Begin()
	}
	if e != nil {
		return existingTx, e
	}
	_, e = existingTx.Exec("DELETE FROM verifiable_address WHERE identity = $1", id)
	if e != nil {
		return existingTx, e
	}
	_, e = existingTx.Exec("DELETE FROM recovery_address WHERE identity = $1", id)
	if e != nil {
		return existingTx, e
	}
	r, e := existingTx.Exec("DELETE FROM identity WHERE id = $1", id)
	if cnt, _ := r.RowsAffected(); e == nil && cnt == 0 {
		e = sql.ErrNoRows
	}
	return existingTx, e
}

//NoRows returns whether error is no rows error
func (s *Store) NoRows(e error) bool {
	return e == sql.ErrNoRows
}

func (s *Store) ensureConnection() (isConnected bool, e error) {
	isConnected = s.db != nil
	if !isConnected {
		e = fmt.Errorf("db connection is not initialized")
	}
	return
}

func (s *Store) insertIdentity(t *sql.Tx, i model.Identity) error {
	_, e := t.Exec("INSERT INTO identity (id, schema_id, schema_url) VALUES ($1, $2, $3)", i.ID, i.SchemaID, i.SchemaURL)
	return e
}

func (s *Store) insertRecoveryAddresses(t *sql.Tx, identity string, a []model.RecoveryAddress) error {
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

func (s *Store) insertVerifiableAddresses(t *sql.Tx, identity string, a []model.VerifiableAddress) error {
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

func (s *Store) findIdentity(id string, identities []model.Identity) (model.Identity, error) {
	for _, i := range identities {
		if i.ID == id {
			return i, nil
		}
	}
	return model.Identity{}, fmt.Errorf("not found")
}

func (s *Store) mapIdentityEntities(identities []model.Identity, verifiableAddresses []model.VerifiableAddress, recoveryAddresses []model.RecoveryAddress) []model.Identity {
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

func (s *Store) execTxChain(operations ...func(*sql.Tx) error) error {
	t, e := s.db.Begin()
	if e != nil {
		return e
	}
	for _, op := range operations {
		if e = op(t); e != nil {
			t.Rollback()
			return e
		}
	}
	return t.Commit()
}
