package main

import (
	"log/slog"
	"net/http"

	"github.com/dllatas/betula/lib"
	"github.com/labstack/echo/v4"
)

type createViewRequest struct {
	Name string   `json:"name"`
	Keys []string `json:"keys"`
	Unit string   `json:"unit"`
}

func (s *Server) createView(c echo.Context) error {
	var req createViewRequest
	if err := c.Bind(&req); err != nil {
		slog.Error("create view: invalid request", "err", err.Error(), "body", req)
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		slog.Error("create view: missing name", "name", req.Name)
		return c.JSON(http.StatusBadRequest, "name is missing")
	}

	unit, err := lib.ParseUnit(req.Unit)
	if err != nil {
		slog.Error("create view: parse unit", "err", err.Error())
		return c.JSON(http.StatusBadRequest, "invalid time unit")
	}

	def := lib.NewViewDefinition(req.Name, req.Keys, unit)
	mapper := lib.NewViewMapper(def)
	instance := lib.NewViewInstance(mapper)

	if err := s.store.Register(instance); err != nil {
		slog.Error("create view: store register", "err", err.Error())
		return c.JSON(http.StatusConflict, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}
