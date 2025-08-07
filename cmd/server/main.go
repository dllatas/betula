package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dllatas/betula/lib"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	store *lib.Store
}

type Config struct {
	DataPath string `default:"./checkpoint.db"`
	Port     string `default:"6357"`
}

func (c Config) Print() {
	slog.Info("betula: current config", "datapath", c.DataPath, "port", c.Port)
}

func main() {
	gob.Register(lib.WALEntry{})
	gob.Register(lib.Event{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var c Config
	err := envconfig.Process("", &c)
	if err != nil {
		logger.Error("betula: failed to parse config", "err", err.Error())
		os.Exit(1)
	}

	c.Print()

	var store *lib.Store
	if c.DataPath == "" {
		logger.Info("betula: running in memory-only mode")
		store = lib.NewInMemoryStore()
	} else {
		logger.Info("betula: running with wal and checkpoint on startup mode")
		store, err = setupStore(c.DataPath)
		if err != nil {
			logger.Error("betula: store setup failed", "err", err.Error())
			os.Exit(2)
		}
	}

	s := &Server{store}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(echoprometheus.NewMiddleware("betula"))

	e.Use(middleware.CORSWithConfig(
		middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{"Authorization", "Content-Type"},
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

	e.GET("/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	v1Group := e.Group("v1")
	v1Group.POST("/views", s.createView)
	v1Group.POST("/append", s.append)
	v1Group.POST("/delete", s.delete)
	v1Group.POST("/spanback", s.spanback)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		bindAddr := fmt.Sprintf(":%s", c.Port)

		if err := e.Start(bindAddr); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("echo server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown: stopping betula server")
		gracefulShutdown(e, store, logger)
	case err := <-errCh:
		logger.Error("fatal error received", "err", err)
		logger.Info("shutdown: stopping betula server after fatal error")
		gracefulShutdown(e, store, logger)
		stop()
	}
}

func gracefulShutdown(e *echo.Echo, store *lib.Store, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}

	if store.WAL != nil {
		if err := store.WAL.Close(); err != nil {
			logger.Error("failed to close WAL", "err", err)
		}
	}
}
