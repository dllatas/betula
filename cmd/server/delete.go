package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/dllatas/betula/lib"
	"github.com/labstack/echo/v4"
)

func (s *Server) delete(c echo.Context) error {
	var req appendRequest

	if err := c.Bind(&req); err != nil {
		slog.Error("delete: failed to parse request body", "err", err.Error())
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}

	if req.View == "" {
		slog.Error("delete: missing view")
		return c.JSON(http.StatusBadRequest, "missing 'view' field")
	}

	if req.Timestamp == "" {
		slog.Error("delete: missing timestamp")
		return c.JSON(http.StatusBadRequest, "missing 'timestamp' field")
	}

	if len(req.Values) == 0 {
		slog.Error("delete: missing keys")
		return c.JSON(http.StatusBadRequest, "missing 'keys' field")
	}

	view, err := s.store.Get(req.View)
	if err != nil {
		slog.Error("delete: view not found", "view", req.View)
		return c.JSON(http.StatusNotFound, err.Error())
	}

	layouts := []string{}
	if req.Timefmt != "" {
		layouts = append(layouts, req.Timefmt)
	}

	layouts = append(layouts, view.Mapper.ViewLayout())

	t := time.Time{}
	for i, layout := range layouts {
		t, err = time.Parse(layout, req.Timestamp)
		if err != nil {
			slog.Error("delete: invalid timestamp", "err", err, "layout", layout, "input", req.Timestamp)

			if i == len(layouts)-1 {
				return c.JSON(http.StatusBadRequest, "invalid timestamp format")
			}

			continue
		}

		break
	}

	e := lib.Event{
		Timestamp: t,
		Labels:    req.Values,
	}

	if s.store.WAL != nil {
		entry := lib.WALEntry{
			Timestamp: time.Now().UTC(),
			Op:        "delete",
			ViewName:  view.Mapper.D.Name,
			Payload:   e,
		}

		if err := s.store.WAL.Write(entry); err != nil {
			slog.Error("delete: failed to write wal", "err", err)
			return c.JSON(http.StatusInternalServerError, "failed to write ahead event")
		}
	}

	if err := view.Delete(e); err != nil {
		slog.Error("delete: failed to delete", "err", err)

		if truncErr := s.store.WAL.RollbackLast(); truncErr != nil {
			slog.Error("delete: failed to rollback wal", "rollbackErr", truncErr)
		}

		return c.JSON(http.StatusInternalServerError, "failed to delete event to view")
	}

	return c.NoContent(http.StatusOK)
}
