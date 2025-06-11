package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/dllatas/betula/lib"
	"github.com/labstack/echo/v4"
)

type appendRequest struct {
	View      string            `json:"view"`
	Timestamp string            `json:"timestamp"`
	Timefmt   string            `json:"time_format,omitempty"`
	Values    map[string]string `json:"keys"`
}

func (s *Server) append(c echo.Context) error {
	var req appendRequest

	if err := c.Bind(&req); err != nil {
		slog.Error("append: failed to parse request body", "err", err.Error())
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}

	if req.View == "" {
		slog.Error("append: missing view")
		return c.JSON(http.StatusBadRequest, "missing 'view' field")
	}

	if req.Timestamp == "" {
		slog.Error("append: missing timestamp")
		return c.JSON(http.StatusBadRequest, "missing 'timestamp' field")
	}

	if len(req.Values) == 0 {
		slog.Error("append: missing keys")
		return c.JSON(http.StatusBadRequest, "missing 'keys' field")
	}

	view, err := s.store.Get(req.View)
	if err != nil {
		slog.Error("append: view not found", "view", req.View)
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
			slog.Error("append: invalid timestamp", "err", err, "layout", layout, "input", req.Timestamp)

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
			Op:        "append",
			ViewName:  view.Mapper.D.Name,
			Payload:   e,
		}

		if err := s.store.WAL.Write(entry); err != nil {
			slog.Error("append: failed to write wal", "err", err)
			return c.JSON(http.StatusInternalServerError, "failed to write ahead event")
		}
	}

	if err := view.Append(e); err != nil {
		slog.Error("append: failed to append", "err", err)

		if truncErr := s.store.WAL.RollbackLast(); truncErr != nil {
			slog.Error("append: failed to rollback wal", "rollbackErr", truncErr)
		}

		return c.JSON(http.StatusInternalServerError, "failed to append event to view")
	}

	return c.NoContent(http.StatusOK)
}
