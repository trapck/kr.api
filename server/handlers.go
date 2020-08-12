package server

import (
	"net/http"

	"github.com/trapck/kr.api/model"

	"github.com/gofiber/fiber"
	uuid "github.com/satori/go.uuid"
)

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
	id := c.Params("id")
	_, err := uuid.FromString(id)
	if err != nil {
		writeError(c, http.StatusBadRequest, err)
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
	id := c.Params("id")
	_, err := uuid.FromString(id)
	if err != nil {
		writeError(c, http.StatusBadRequest, err)
		return
	}
	err = a.store.Delete(id)
	if err != nil {
		writeError(c, a.statusFromDBErr(err), err)
		return
	}
	c.Status(http.StatusNoContent)
}

//HandleCreate handles create identitiy request
func (a *IdentApp) HandleCreate(c *fiber.Ctx) {
	var i model.Identity
	err := c.BodyParser(&i)
	//TODO: validate json struct
	if err != nil {
		writeError(c, http.StatusBadRequest, err)
		return
	}
	i, err = a.store.Create(i)
	if err != nil {
		writeError(c, a.statusFromDBErr(err), err)
		return
	}
	writeSuccess(c, http.StatusCreated, i)
}

//HandleUpdate handles update identitiy request
func (a *IdentApp) HandleUpdate(c *fiber.Ctx) {
	id := c.Params("id")
	_, err := uuid.FromString(id)
	if err != nil {
		writeError(c, http.StatusBadRequest, err)
		return
	}
	var i model.Identity
	err = c.BodyParser(&i)
	//TODO: validate json struct
	if err != nil {
		writeError(c, http.StatusBadRequest, err)
		return
	}
	i, err = a.store.Update(id, i)
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
