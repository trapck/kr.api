package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/trapck/kr.api/model"

	"github.com/gofiber/fiber"
	uuid "github.com/satori/go.uuid"
	"github.com/xeipuuv/gojsonschema"
)

var identityJSONSchema = gojsonschema.NewReferenceLoader("file:///home/trapck/Desktop/krapi/model/schema.json")

//HandleList handles list all identities request
func (a *IdentApp) HandleList(c *fiber.Ctx) {
	l, err := a.store.List()
	if err != nil {
		writeError(c, a.statusFromDBErr(err), err)
		return
	}
	writeSuccess(c, http.StatusOK, l)
}

//HandleGet handles get identitiy request
func (a *IdentApp) HandleGet(c *fiber.Ctx) {
	id, valid := extractIDParam(c)
	if !valid {
		return
	}
	i, err := a.store.Get(id)
	if err != nil {
		writeError(c, a.statusFromDBErr(err), err)
		return
	}
	writeSuccess(c, http.StatusOK, i)
}

//HandleDelete handles delete identitiy request
func (a *IdentApp) HandleDelete(c *fiber.Ctx) {
	id, valid := extractIDParam(c)
	if !valid {
		return
	}
	err := a.store.Delete(id)
	if err != nil {
		writeError(c, a.statusFromDBErr(err), err)
		return
	}
	c.Status(http.StatusNoContent)
}

//HandleCreate handles create identitiy request
func (a *IdentApp) HandleCreate(c *fiber.Ctx) {
	var i model.Identity
	if !parseIdentity(c, &i) {
		return
	}
	i, err := a.store.Create(i)
	if err != nil {
		writeError(c, a.statusFromDBErr(err), err)
		return
	}
	writeSuccess(c, http.StatusCreated, i)
}

//HandleUpdate handles update identitiy request
func (a *IdentApp) HandleUpdate(c *fiber.Ctx) {
	id, valid := extractIDParam(c)
	if !valid {
		return
	}
	var i model.Identity
	if !parseIdentity(c, &i) {
		return
	}
	i, err := a.store.Update(id, i)
	if err != nil {
		writeError(c, a.statusFromDBErr(err), err)
		return
	}
	writeSuccess(c, http.StatusOK, i)
}

func writeSuccess(c *fiber.Ctx, code int, data interface{}) {
	c.JSON(data)
	c.Status(code)
}

func writeError(c *fiber.Ctx, code int, e error) {
	c.JSON(model.NewGenericErrorWrap(code, e))
	c.Status(code)
}

func (a *IdentApp) statusFromDBErr(e error) int {
	if a.store.NoRows(e) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func validateIdentityJSON(src string) (*gojsonschema.Result, error) {
	return gojsonschema.Validate(identityJSONSchema, gojsonschema.NewStringLoader(src))
}

func combineJSONSchemaErrors(e []gojsonschema.ResultError) error {
	s := make([]string, len(e))
	for i, v := range e {
		s[i] = v.String()
	}
	return fmt.Errorf(strings.Join(s, "\n"))
}

func parseIdentity(c *fiber.Ctx, i *model.Identity) bool {
	r, err := validateIdentityJSON(c.Body())
	if err != nil {
		writeError(c, http.StatusBadRequest, err)
		return false
	} else if !r.Valid() {
		writeError(c, http.StatusUnprocessableEntity, combineJSONSchemaErrors(r.Errors()))
		return false
	}
	if err != c.BodyParser(&i) {
		writeError(c, http.StatusBadRequest, err)
		return false
	}
	return true
}

func extractIDParam(c *fiber.Ctx) (string, bool) {
	id := c.Params("id")
	_, err := uuid.FromString(id)
	if err != nil {
		writeError(c, http.StatusBadRequest, err)
		return id, false
	}
	return id, true
}
