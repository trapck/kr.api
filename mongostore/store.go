package mongostore

import (
	"context"
	"fmt"
	"time"

	"github.com/trapck/kr.api/appconfig"
	"github.com/trapck/kr.api/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Store is mongodb storage implementation
type Store struct {
	client   *mongo.Client
	db       *mongo.Database
	identity *mongo.Collection
}

// Init initializes connetion
func (s *Store) Init() error {
	ctx, cancel := ctx()
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(appconfig.MongoHost))
	if err != nil {
		return fmt.Errorf("Unable to connect to mongodb: %+v", err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("Mongo ping failed: %+v", err)
	}
	s.client = client
	s.db = client.Database(appconfig.MongoDBName)
	err = s.initDocuments(ctx)
	if err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("Error when initializing on of db documents %+v", err)
	}
	return nil
}

// Close closes connetion
func (s *Store) Close() error {
	ctx, cancel := ctx()
	defer cancel()
	return s.client.Disconnect(ctx)
}

// List returns all identities
func (s *Store) List() ([]model.Identity, error) {
	ctx, cancel := ctx()
	defer cancel()
	cur, err := s.identity.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	res := []model.Identity{}
	for cur.Next(ctx) {
		var i model.Identity
		err := cur.Decode(&i)
		if err != nil {
			return nil, err
		}
		res = append(res, i)
	}
	if err = cur.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// Get returns all identities
func (s *Store) Get(id string) (model.Identity, error) {
	ctx, cancel := ctx()
	defer cancel()
	var i model.Identity
	r := s.identity.FindOne(ctx, idFilter(id))
	if err := r.Err(); err != nil {
		return i, err
	}
	err := r.Decode(&i)

	return i, err
}

// Create inserts identity
func (s *Store) Create(i model.Identity) (model.Identity, error) {
	ctx, cancel := ctx()
	defer cancel()
	_, err := s.identity.InsertOne(ctx, i)
	return i, err
}

// Update updates identity
func (s *Store) Update(id string, i model.Identity) (model.Identity, error) {
	ctx, cancel := ctx()
	defer cancel()
	r, err := s.identity.ReplaceOne(ctx, idFilter(id), i)
	if err == nil && r.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return i, err
}

//Delete deletes identity
func (s *Store) Delete(id string) error {
	ctx, cancel := ctx()
	defer cancel()
	r, err := s.identity.DeleteOne(ctx, idFilter(id))
	if err == nil && r.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	return err
}

//NoRows returns whether error is no rows error
func (s *Store) NoRows(e error) bool {
	return e == mongo.ErrNoDocuments
}

func (s *Store) initDocuments(ctx context.Context) error {
	s.identity = s.db.Collection("identity")
	_, err := s.identity.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
	)
	return err
}

func ctx(timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := appconfig.MongoDefTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(context.Background(), t)
}

func idFilter(id string) bson.M {
	return bson.M{"id": id}
}
