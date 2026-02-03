package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dllatas/betula/lib"
	"github.com/labstack/echo/v4"
)

func TestCreateViewSuccess(t *testing.T) {
	e := echo.New()
	s := &Server{store: lib.NewStore()}

	req := httptest.NewRequest(http.MethodPost, "/v1/views", strings.NewReader(`{"name":"sessions","keys":["user"],"unit":"minute"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	if err := s.createView(e.NewContext(req, rec)); err != nil {
		t.Fatalf("createView returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if _, err := s.store.Get("sessions"); err != nil {
		t.Fatalf("view was not stored: %v", err)
	}
}

func TestCreateViewInvalidUnit(t *testing.T) {
	e := echo.New()
	s := &Server{store: lib.NewStore()}

	req := httptest.NewRequest(http.MethodPost, "/v1/views", strings.NewReader(`{"name":"sessions","keys":["user"],"unit":"invalid"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	if err := s.createView(e.NewContext(req, rec)); err != nil {
		t.Fatalf("createView returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestAppendValidationErrors(t *testing.T) {
	e := echo.New()
	s := &Server{store: lib.NewStore()}

	// Missing view should return 404
	req := httptest.NewRequest(http.MethodPost, "/v1/append", strings.NewReader(`{"view":"missing","timestamp":"2024-01-01 00:00","keys":{"user":"u1"}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	if err := s.append(e.NewContext(req, rec)); err != nil {
		t.Fatalf("append returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	// Invalid timestamp should return 400
	viewName := "append-missing"
	registerTestView(t, s, viewName, []string{"user"}, lib.UnitMinute)
	reqBadTime := httptest.NewRequest(http.MethodPost, "/v1/append", strings.NewReader(`{"view":"`+viewName+`","timestamp":"bad time","keys":{"user":"u1"}}`))
	reqBadTime.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recBadTime := httptest.NewRecorder()

	if err := s.append(e.NewContext(reqBadTime, recBadTime)); err != nil {
		t.Fatalf("append returned error: %v", err)
	}
	if recBadTime.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recBadTime.Code)
	}
}

func TestAppendAndSpanbackFlow(t *testing.T) {
	e := echo.New()
	s := &Server{store: lib.NewStore()}
	viewName := "metrics"
	view := registerTestView(t, s, viewName, []string{"country", "user"}, lib.UnitMinute)

	ts := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC).Format(view.Mapper.Layout())
	appendBody := map[string]any{
		"view":      viewName,
		"timestamp": ts,
		"keys": map[string]string{
			"country": "fr",
			"user":    "u1",
		},
	}
	appendPayload, _ := json.Marshal(appendBody)

	appendReq := httptest.NewRequest(http.MethodPost, "/v1/append", bytes.NewReader(appendPayload))
	appendReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	appendRec := httptest.NewRecorder()

	if err := s.append(e.NewContext(appendReq, appendRec)); err != nil {
		t.Fatalf("append returned error: %v", err)
	}
	if appendRec.Code != http.StatusOK {
		t.Fatalf("expected append status %d, got %d", http.StatusOK, appendRec.Code)
	}

	spanbackBody := map[string]any{
		"view":     viewName,
		"unit":     "minute",
		"value":    10,
		"ref":      ts,
		"filter":   map[string][]string{"country": {"fr"}},
		"group_by": []string{"country"},
	}
	spanbackPayload, _ := json.Marshal(spanbackBody)

	spanbackReq := httptest.NewRequest(http.MethodPost, "/v1/spanback", bytes.NewReader(spanbackPayload))
	spanbackReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	spanbackRec := httptest.NewRecorder()

	if err := s.spanback(e.NewContext(spanbackReq, spanbackRec)); err != nil {
		t.Fatalf("spanback returned error: %v", err)
	}
	if spanbackRec.Code != http.StatusOK {
		t.Fatalf("expected spanback status %d, got %d", http.StatusOK, spanbackRec.Code)
	}

	var grouped map[string]int
	if err := json.Unmarshal(spanbackRec.Body.Bytes(), &grouped); err != nil {
		t.Fatalf("failed to unmarshal spanback response: %v", err)
	}
	if grouped["fr"] != 1 {
		t.Fatalf("expected grouped count 1 for 'fr', got %d", grouped["fr"])
	}
}

func TestSpanbackValidationErrors(t *testing.T) {
	e := echo.New()
	s := &Server{store: lib.NewStore()}
	viewName := "spanback"
	registerTestView(t, s, viewName, []string{"user"}, lib.UnitMinute)

	// Missing required fields
	req := httptest.NewRequest(http.MethodPost, "/v1/spanback", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	if err := s.spanback(e.NewContext(req, rec)); err != nil {
		t.Fatalf("spanback returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Invalid ref format
	body := map[string]any{
		"view":  viewName,
		"unit":  "minute",
		"value": 1,
		"ref":   "bad",
	}
	payload, _ := json.Marshal(body)
	reqBad := httptest.NewRequest(http.MethodPost, "/v1/spanback", bytes.NewReader(payload))
	reqBad.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recBad := httptest.NewRecorder()

	if err := s.spanback(e.NewContext(reqBad, recBad)); err != nil {
		t.Fatalf("spanback returned error: %v", err)
	}
	if recBad.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recBad.Code)
	}
}

func registerTestView(t *testing.T, s *Server, name string, keys []string, unit lib.TimeUnit) *lib.ViewInstance {
	t.Helper()
	def := lib.NewViewDefinition(name, keys, unit)
	mapper := lib.NewViewMapper(def)
	instance := lib.NewViewInstance(mapper)

	if err := s.store.Register(instance); err != nil {
		t.Fatalf("failed to register test view: %v", err)
	}

	return instance
}
