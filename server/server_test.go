package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/trapck/kr.api/model"
	"github.com/trapck/kr.api/testutil"
)

const notFound = "not found"

type stubStore struct {
	identities []model.Identity
}

func (s *stubStore) List() ([]model.Identity, error) {
	return s.identities, nil
}

func (s *stubStore) Get(id string) (model.Identity, error) {
	var r model.Identity
	e := fmt.Errorf(notFound)
	for _, v := range s.identities {
		if v.ID == id {
			r = v
			e = nil
			break
		}
	}
	return r, e
}

func (s *stubStore) Create(i model.Identity) (model.Identity, error) {
	s.identities = append(s.identities, i)
	return i, nil
}

func (s *stubStore) Update(id string, i model.Identity) (model.Identity, error) {
	e := s.Delete(id)
	if e != nil {
		return model.Identity{}, e
	}
	return s.Create(i)
}

func (s *stubStore) Delete(id string) error {
	var pos int
	e := fmt.Errorf(notFound)
	for i, v := range s.identities {
		if v.ID == id {
			pos = i
			e = nil
			break
		}
	}
	if e != nil {
		return e
	}
	s.identities = append(s.identities[:pos], s.identities[pos+1:]...)
	return nil
}

func (s *stubStore) NoRows(e error) bool {
	return e.Error() == notFound
}

func TestList(t *testing.T) {
	store := stubStore{identities: []model.Identity{model.Identity{ID: uuid.NewV4().String()}, model.Identity{ID: uuid.NewV4().String()}}}
	srv := NewApp(&store)
	req, _ := http.NewRequest(http.MethodGet, "/identities", nil)
	resp, err := srv.server.Test(req)
	testutil.FailOnNotEqual(t, err, nil, fmt.Sprintf("got an error while serving http request %v", err))
	body := []model.Identity{}
	assertSussessJSONResponse(t, http.StatusOK, resp, &body)
	assert.Equal(t, store.identities, body, "response list doesnt match store list")
}

func TestGet(t *testing.T) {
	id := uuid.NewV4().String()
	store := stubStore{identities: []model.Identity{model.Identity{ID: id}}}
	srv := NewApp(&store)
	t.Run("should return existing identity", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/identities/"+id, nil)
		resp, err := srv.server.Test(req)
		testutil.FailOnNotEqual(t, err, nil, fmt.Sprintf("got an error while serving http request %v", err))
		body := model.Identity{}
		assertSussessJSONResponse(t, http.StatusOK, resp, &body)
		assert.Equal(t, store.identities[0], body, "response identity doesnt match store identity")
	})
	t.Run("should return bad request for invalid id format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/identities/1", nil)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusBadRequest, resp)
	})
	t.Run("should return not found for not existing id", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/identities/"+uuid.NewV4().String(), nil)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusNotFound, resp)
	})
}

func TestCreate(t *testing.T) {
	id := uuid.NewV4().String()
	toCreate := model.Identity{ID: id}
	store := stubStore{}
	srv := NewApp(&store)

	t.Run("should create identity", func(t *testing.T) {
		serializedIdentity, _ := json.Marshal(toCreate)
		req, _ := http.NewRequest(http.MethodPost, "/identities", bytes.NewBuffer(serializedIdentity))
		req.Header.Set(HeaderKeyContentType, HeaderValueJSONContactType)
		resp, err := srv.server.Test(req)
		testutil.FailOnNotEqual(t, err, nil, fmt.Sprintf("got an error while serving http request %v", err))
		body := model.Identity{}
		assertSussessJSONResponse(t, http.StatusCreated, resp, &body)
		assert.Equal(t, toCreate, body, "response identity doesnt match request identity")
		assert.Equal(t, store.identities[0], body, "response identity doesnt match store identity")
	})
	t.Run("should return bad request for invalid json", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/identities", bytes.NewBuffer([]byte("{")))
		req.Header.Set(HeaderKeyContentType, HeaderValueJSONContactType)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusBadRequest, resp)
	})
}

func TestUpdate(t *testing.T) {
	existingID := uuid.NewV4().String()
	newID := uuid.NewV4().String()
	toUpdate := model.Identity{ID: existingID}
	newData := model.Identity{ID: newID}
	store := stubStore{identities: []model.Identity{toUpdate}}
	srv := NewApp(&store)
	serializedIdentity, _ := json.Marshal(newData)

	t.Run("should uodate existing identity", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/identities/"+existingID, bytes.NewBuffer(serializedIdentity))
		req.Header.Set(HeaderKeyContentType, HeaderValueJSONContactType)
		resp, err := srv.server.Test(req)
		testutil.FailOnNotEqual(t, err, nil, fmt.Sprintf("got an error while serving http request %v", err))
		body := model.Identity{}
		assertSussessJSONResponse(t, http.StatusOK, resp, &body)
		assert.Equal(t, newData, body, "response identity doesnt match request identity")
		assert.Equal(t, store.identities[0], body, "response identity doesnt match store identity")
	})
	t.Run("should return bad request for invalid id format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/identities/1", bytes.NewBuffer(serializedIdentity))
		req.Header.Set(HeaderKeyContentType, HeaderValueJSONContactType)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusBadRequest, resp)
	})
	t.Run("should return bad request for invalid json", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/identities/"+existingID, bytes.NewBuffer([]byte("{")))
		req.Header.Set(HeaderKeyContentType, HeaderValueJSONContactType)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusBadRequest, resp)
	})
	t.Run("should return not found for not existing id", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/identities/"+uuid.NewV4().String(), bytes.NewBuffer(serializedIdentity))
		req.Header.Set(HeaderKeyContentType, HeaderValueJSONContactType)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusNotFound, resp)
	})
}

func TestDelete(t *testing.T) {
	id := uuid.NewV4().String()
	store := stubStore{identities: []model.Identity{model.Identity{ID: id}}}
	srv := NewApp(&store)

	t.Run("should delete existing identity", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/identities/"+id, nil)
		resp, err := srv.server.Test(req)
		testutil.FailOnNotEqual(t, err, nil, fmt.Sprintf("got an error while serving http request %v", err))
		assertStatus(t, http.StatusNoContent, resp.StatusCode, "")
		assert.Equal(t, len(store.identities), 0, "identity was not deleted from store")
	})
	t.Run("should return bad request for invalid id format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/identities/1", nil)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusBadRequest, resp)
	})
	t.Run("should return not found for not existing id", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/identities/"+uuid.NewV4().String(), nil)
		resp, _ := srv.server.Test(req)
		assertErrorJSONResponse(t, http.StatusNotFound, resp)
	})
}

func assertStatus(t *testing.T, want, got int, message string) {
	t.Helper()
	testutil.FailOnNotEqual(t, want, got, fmt.Sprintf("didn't get correct status. got %d instead of %d. %s", got, want, message))
}

func assertJSONContentType(t *testing.T, resp *http.Response) {
	t.Helper()
	got, want := resp.Header.Get(HeaderKeyContentType), HeaderValueJSONContactType
	testutil.FailOnNotEqual(t, want, got, fmt.Sprintf("invalid content-type. got %q instead of %q", got, want))
}

func assertSuccessJSONResponseHeaders(t *testing.T, code int, resp *http.Response) {
	t.Helper()
	assertStatus(t, code, resp.StatusCode, "on valid request")
	assertJSONContentType(t, resp)
}

func assertJSONBody(t *testing.T, gotJSON string, compareTo interface{}, msg string) {
	t.Helper()
	builder := strings.Builder{}
	json.NewEncoder(&builder).Encode(compareTo)
	assert.Equal(t, builder.String(), gotJSON, msg)
}

func assertSussessJSONResponse(t *testing.T, code int, resp *http.Response, decodeTo interface{}) {
	t.Helper()
	assertSuccessJSONResponseHeaders(t, code, resp)
	err := json.NewDecoder(resp.Body).Decode(decodeTo)
	b, _ := ioutil.ReadAll(resp.Body)
	testutil.FailOnNotEqual(t, err, nil, fmt.Sprintf("unable to decode %q to type %T. Got error %q", string(b), decodeTo, err))
}

func assertErrorJSONResponse(t *testing.T, code int, resp *http.Response) {
	t.Helper()
	assertStatus(t, code, resp.StatusCode, "on broken request")
	assertJSONContentType(t, resp)
	var e model.GenericErrorWrap
	json.NewDecoder(resp.Body).Decode(&e)
	testutil.FailOnEqual(t, e.Error.Message, "", "haven't got and error description")
}
