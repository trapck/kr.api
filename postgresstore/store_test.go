package postgresstore

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/trapck/kr.api/model"
	"github.com/trapck/kr.api/testutil"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	db := initDB(t)
	defer closeDB(t, db)
	expected := []model.Identity{}
	db.db.Select(&expected, "SELECT * FROM identity")
	actual, err := db.List()
	testutil.FailOnNotEqual(t, err, nil, "expected db operation to be finished with no errors")
	assert.Equal(t, len(expected), len(actual), "result list count doesn't match desired count")
}

func TestCreate(t *testing.T) {
	db := initDB(t)
	sessionID := createSessionID()
	time := time.Now()
	defer closeDB(t, db)
	defer clearAllTestData(db, sessionID)

	input := model.Identity{
		ID: uuid.NewV4().String(),
		RecoveryAddresses: []model.RecoveryAddress{
			model.RecoveryAddress{
				Address: model.Address{
					ID:    uuid.NewV4().String(),
					Value: sessionID,
					Via:   "via2",
				},
				Identity: "2f6aa0ab-d2a3-4adf-ae48-9183bfa26389",
			},
		},
		SchemaID:  sessionID,
		SchemaURL: "2.com",
		VerifiableAddresses: []model.VerifiableAddress{
			model.VerifiableAddress{
				Address: model.Address{
					ID:    uuid.NewV4().String(),
					Value: sessionID,
					Via:   "via2",
				},
				Verified:   true,
				VerifiedAt: &time,
				ExpiresAt:  &time,
				Identity:   "2f6aa0ab-d2a3-4adf-ae48-9183bfa26389",
			},
		},
	}
	output, err := db.Create(input)
	testutil.FailOnNotEqual(t, err, nil, fmt.Sprintf("must be created without error, instead got : %s", err))
	assert.Equal(t, input, output, fmt.Sprintf("created identity must be equal to input"))
	var found model.Identity
	err = db.db.Get(&found, "SELECT * FROM identity WHERE id = $1", output.ID)
	assert.NoError(t, err, fmt.Sprintf("expected to find identity in db"))
}

func TestGet(t *testing.T) {
	db := initDB(t)
	sessionID := createSessionID()
	id := uuid.NewV4().String()
	defer closeDB(t, db)
	defer clearAllTestData(db, sessionID)

	data := model.Identity{
		ID:                  id,
		SchemaID:            sessionID,
		RecoveryAddresses:   []model.RecoveryAddress{model.RecoveryAddress{Address: model.Address{ID: id, Value: sessionID}, Identity: id}},
		VerifiableAddresses: []model.VerifiableAddress{model.VerifiableAddress{Address: model.Address{ID: id, Value: sessionID}, Identity: id}},
	}

	db.db.Exec("INSERT INTO identity (id, schema_id) VALUES ($1, $2)", id, sessionID)
	db.db.Exec("INSERT INTO recovery_address (id, value, identity) VALUES ($1, $2, $3)", id, sessionID, id)
	db.db.Exec("INSERT INTO verifiable_address (id, value, identity) VALUES ($1, $2, $3)", id, sessionID, id)
	found, err := db.Get(id)
	testutil.FailOnNotEqual(t, err, nil, "expected to get identity without errors")
	assert.Equal(t, data, found, "expected to find identity in db with desired struct")
}

func TestDelete(t *testing.T) {
	db := initDB(t)
	sessionID := createSessionID()
	id := uuid.NewV4().String()
	defer closeDB(t, db)
	defer clearAllTestData(db, sessionID)

	db.db.Exec("INSERT INTO identity (id, schema_id) VALUES ($1, $2)", id, sessionID)
	db.db.Exec("INSERT INTO recovery_address (id, value, identity) VALUES ($1, $2, $3)", id, sessionID, id)
	db.db.Exec("INSERT INTO verifiable_address (id, value, identity) VALUES ($1, $2, $3)", id, sessionID, id)
	err := db.Delete(id)
	testutil.FailOnNotEqual(t, err, nil, "expected to delete identity without error")
	found := model.Identity{}
	err = db.db.Get(&found, "SELECT * FROM identity WHERE id = $1", id)
	assert.Error(t, err, "expected identity to be not found in db")
}

func initDB(t *testing.T) *Store {
	t.Helper()
	db := Store{}
	err := db.Init()
	if err != nil {
		assert.FailNow(t, "db connection was not established. ", err)
	}
	return &db
}

func closeDB(t *testing.T, db *Store) {
	t.Helper()
	err := db.Close()
	if err != nil {
		assert.FailNow(t, "db connection was not closed. ", err)
	}
}

func clearTestData(db *Store, table, filter string) (sql.Result, error) {
	return db.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s", table, filter))
}

func createSessionID() string {
	return strconv.Itoa(int(rand.Uint32()))
}

func clearAllTestData(db *Store, sessionID string) {
	clearTestData(db, "recovery_address", fmt.Sprintf("value='%s'", sessionID))
	clearTestData(db, "verifiable_address", fmt.Sprintf("value='%s'", sessionID))
	clearTestData(db, "identity", fmt.Sprintf("schema_id='%s'", sessionID))
}
