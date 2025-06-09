package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dllatas/betula/lib"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	store *lib.Store
}

func main() {
	ctx := context.Background()

	store := lib.NewStore()
	s := &Server{store}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(echoprometheus.NewMiddleware("betula"))

	e.Use(middleware.CORSWithConfig(
		middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		},
	))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:       true,
		LogLatency:      true,
		LogMethod:       true,
		LogURI:          true,
		LogRoutePath:    true,
		LogRequestID:    true,
		LogRemoteIP:     true,
		LogProtocol:     true,
		LogUserAgent:    true,
		LogReferer:      true,
		LogResponseSize: true,
		LogError:        true,
		HandleError:     true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.String("route", v.RoutePath),
				slog.String("request_id", v.RequestID),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("protocol", v.Protocol),
				slog.String("user_agent", v.UserAgent),
				slog.String("referer", v.Referer),
				slog.Int("status", v.Status),
				slog.String("latency", v.Latency.String()),
				slog.Int64("response_size", v.ResponseSize),
			}

			if v.Error != nil {
				attrs = append(attrs, slog.String("err", v.Error.Error()))
				logger.LogAttrs(
					c.Request().Context(),
					slog.LevelError,
					"ERROR",
					attrs...,
				)
				return nil
			}

			logger.LogAttrs(
				c.Request().Context(),
				slog.LevelInfo,
				"REQUEST",
				attrs...,
			)

			return nil
		},
	}))

	e.GET("/metrics", echoprometheus.NewHandler())

	v1Group := e.Group("v1")
	v1Group.POST("/views", s.createView)
	v1Group.POST("/append", s.append)
	v1Group.POST("/spanback", s.spanback)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	go func() {
		if err := e.Start(":6357"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("betula: shutting down the server")
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
