package main

import (
	"log/slog"
	"net/http"
	"time"

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

	if s.store.WAL != nil {
		entry := lib.WALEntry{
			Timestamp: time.Now().UTC(),
			Op:        "create-view",
			ViewName:  req.Name,
			Keys:      req.Keys,
			Unit:      req.Unit,
		}

		if err := s.store.WAL.Write(entry); err != nil {
			slog.Error("delete: failed to write wal", "err", err)
			return c.JSON(http.StatusInternalServerError, "failed to write ahead event")
		}
	}

	if err := s.store.Register(instance); err != nil {
		slog.Error("create view: store register", "err", err.Error())

		if truncErr := s.store.WAL.RollbackLast(); truncErr != nil {
			slog.Error("create view: failed to rollback wal", "rollbackErr", truncErr)
		}

		return c.JSON(http.StatusConflict, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}
