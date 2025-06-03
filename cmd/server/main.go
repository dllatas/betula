package main

import (
	"github.com/dllatas/betula/lib"
	"github.com/labstack/echo/v4"
)

type Server struct {
	store *lib.Store
}

func main() {
	store := lib.NewStore()

	s := &Server{store}

	e := echo.New()
	v1Group := e.Group("v1")

	v1Group.POST("/views", s.createView)
	v1Group.POST("/append", s.append)
	v1Group.POST("/spanback", s.spanback)

	e.Logger.Fatal(e.Start(":6357"))
}
