package mongostore

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	uuid "github.com/satori/go.uuid"
	"github.com/trapck/kr.api/model"
	"github.com/trapck/kr.api/testutil"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	db := initDB(t)
	defer closeDB(t, db)
	context, cancel := ctx()
	defer cancel()
	cnt, err := db.identity.CountDocuments(context, bson.M{})
	testutil.FailOnNotEqual(t, err, nil, "error when counting identities for comparison")
	actual, err := db.List()
	testutil.FailOnNotEqual(t, err, nil, "expected db operation to be finished with no errors")
	assert.Equal(t, int(cnt), len(actual), "result list count doesn't match desired count")
}

func TestCreate(t *testing.T) {
	sessionID := createSessionID()
	time := time.Now()
	db := initDB(t)
	defer closeDB(t, db)
	context, cancel := ctx()
	defer cancel()
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
	cnt, err := db.identity.CountDocuments(context, idFilter(output.ID))
	testutil.FailOnNotEqual(t, err, nil, "error when counting identities for comparison")
	assert.NotEqual(t, 0, cnt, fmt.Sprintf("expected to find identity in db"))
}

func TestGet(t *testing.T) {
	sessionID := createSessionID()
	id := uuid.NewV4().String()
	db := initDB(t)
	defer closeDB(t, db)
	context, cancel := ctx()
	defer cancel()
	defer clearAllTestData(db, sessionID)

	data := model.Identity{
		ID:                  id,
		SchemaID:            sessionID,
		RecoveryAddresses:   []model.RecoveryAddress{model.RecoveryAddress{Address: model.Address{ID: id, Value: sessionID}, Identity: id}},
		VerifiableAddresses: []model.VerifiableAddress{model.VerifiableAddress{Address: model.Address{ID: id, Value: sessionID}, Identity: id}},
	}
	_, err := db.identity.InsertOne(context, data)
	testutil.FailOnNotEqual(t, err, nil, "error when inserting identity for comparison")

	t.Run("should return an existing entity", func(t *testing.T) {
		found, err := db.Get(id)
		testutil.FailOnNotEqual(t, err, nil, "expected to get identity without errors")
		assert.Equal(t, data, found, "expected to find identity in db with desired struct")
	})
	t.Run("should return an error for not existing entity", func(t *testing.T) {
		_, err := db.Get(uuid.NewV4().String())
		assert.Error(t, err, "expected to get an error for not existing identity select")
	})
}
func TestUpdate(t *testing.T) {
	sessionID := createSessionID()
	db := initDB(t)
	defer closeDB(t, db)
	context, cancel := ctx()
	defer cancel()
	defer clearAllTestData(db, sessionID)

	oldData := model.Identity{ID: uuid.NewV4().String(), SchemaID: sessionID}
	newData := model.Identity{ID: uuid.NewV4().String(), SchemaID: sessionID}
	_, err := db.identity.InsertOne(context, oldData)
	testutil.FailOnNotEqual(t, err, nil, "error when inserting identity for comparison")

	t.Run("should update existing entity", func(t *testing.T) {
		_, err = db.Update(oldData.ID, newData)
		testutil.FailOnNotEqual(t, err, nil, "expected to update identity without errors")
		cnt, err := db.identity.CountDocuments(context, idFilter(newData.ID))
		testutil.FailOnNotEqual(t, err, nil, "error when counting identities for comparison")
		assert.NotEqual(t, 0, cnt, fmt.Sprintf("expected to find identity with new data in db"))

	})
	t.Run("should return error for not existing entity", func(t *testing.T) {
		_, err = db.Update(uuid.NewV4().String(), newData)
		assert.Error(t, err, "expected to get an error for not existing identity update")
	})
}

func TestDelete(t *testing.T) {
	sessionID := createSessionID()
	id := uuid.NewV4().String()
	db := initDB(t)
	defer closeDB(t, db)
	context, cancel := ctx()
	defer cancel()
	defer clearAllTestData(db, sessionID)

	_, err := db.identity.InsertOne(context, model.Identity{ID: id, SchemaID: sessionID})
	testutil.FailOnNotEqual(t, err, nil, "error when inserting identity for comparison")
	t.Run("should delete an existing entity", func(t *testing.T) {
		err = db.Delete(id)
		testutil.FailOnNotEqual(t, err, nil, "expected to delete identity without error")
		cnt, err := db.identity.CountDocuments(context, idFilter(id))
		testutil.FailOnNotEqual(t, err, nil, "expected to get identity without errors")
		assert.Equal(t, 0, int(cnt), "expected identity to be not found in db")
	})
	t.Run("should return error for not existing entity", func(t *testing.T) {
		err = db.Delete(uuid.NewV4().String())
		assert.Error(t, err, "expected to get an error for not existing identity delete")
	})
}

func initDB(t *testing.T) *Store {
	t.Helper()
	s := Store{}
	if err := s.Init(); err != nil {
		assert.FailNow(t, "db connection was not established. ", err)
	}
	return &s
}

func closeDB(t *testing.T, db *Store) {
	t.Helper()
	err := db.Close()
	if err != nil {
		assert.FailNow(t, "db connection was not closed. ", err)
	}
}

func createSessionID() string {
	return strconv.Itoa(int(rand.Uint32()))
}

func clearAllTestData(db *Store, sessionID string) {
	context, cancel := ctx()
	defer cancel()
	db.identity.DeleteMany(context, bson.M{"schema_id": primitive.Regex{Pattern: sessionID + "$", Options: ""}})
}
