package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/dllatas/betula/lib"
	"github.com/labstack/echo/v4"
)

type SpanbackRequest struct {
	View    string              `json:"view"`
	Ref     string              `json:"ref,omitempty"` // optional
	Unit    string              `json:"unit"`          // required
	Value   int                 `json:"value"`         // required
	Verbose bool                `json:"verbose,omitempty"`
	Filter  map[string][]string `json:"filter,omitempty"`
	GroupBy []string            `json:"group_by,omitempty"`
	Timefmt string              `json:"time_format,omitempty"` // optional
}

func (s *Server) spanback(c echo.Context) error {
	var req SpanbackRequest

	if err := c.Bind(&req); err != nil {
		slog.Error("spanback: bind failed", "err", err)
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}

	if req.View == "" || req.Unit == "" || req.Value <= 0 {
		slog.Error("spanback: bind failed", "view", req.View, "unit", req.Unit, "value", req.Value)
		return c.JSON(http.StatusBadRequest, "view, unit, and value are required")
	}

	view, err := s.store.Get(req.View)
	if err != nil {
		slog.Error("append: view not found", "view", req.View)
		return c.JSON(http.StatusNotFound, err.Error())
	}

	unit, err := lib.ParseUnit(req.Unit)
	if err != nil {
		slog.Error("append: parse unit failed", "unit", req.Unit)
		return c.JSON(http.StatusBadRequest, "invalid time unit")
	}

	ref := time.Now()
	if req.Ref != "" {
		layout := req.Timefmt
		if layout == "" {
			layout = view.Mapper.Layout()
		}
		ref, err = time.Parse(layout, req.Ref)
		if err != nil {
			slog.Error("append: parse time failed", "layout", layout, "ref", req.Ref)
			return c.JSON(http.StatusBadRequest, "invalid ref timestamp")
		}
	}

	shards, err := view.SpanBack(ref, unit, req.Value, req.Verbose)
	if err != nil {
		slog.Error("spanback: spanback failed", "err", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if len(req.Filter) > 0 {
		var errFilter error
		shards, errFilter = view.FilterWithShards(req.Filter, shards)
		if errFilter != nil {
			slog.Error("spanback: filter failed", "err", errFilter)
			return c.JSON(http.StatusInternalServerError, errFilter.Error())
		}
	}

	if len(req.GroupBy) > 0 {
		result, err := view.GroupByWithShards(req.GroupBy, shards)
		if err != nil {
			slog.Error("spanback: groupby failed", "err", err)
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, result)
	}

	return c.JSON(http.StatusOK, shards)
}
