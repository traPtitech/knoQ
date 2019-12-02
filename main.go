package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	repo "room/repository"
	"room/router"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db, err := repo.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// echo初期化
	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			if he.Internal != nil {
				if herr, ok := he.Internal.(*echo.HTTPError); ok {
					fmt.Println(herr)
					he = herr
				}
			}
		} else {
			he = &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			}
		}
		c.Set("Error-Runtime", he.Internal)

		// Issue #1426
		code := he.Code
		message := he.Message
		if e.Debug {
			message = err.Error()
		} else if m, ok := message.(string); ok {
			message = echo.Map{"message": m}
		}

		// Send response
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				err = c.NoContent(he.Code)
			} else {
				err = c.JSON(code, message)
			}
			if err != nil {
				e.Logger.Error(err)
			}
		}
	}
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "web/dist",
		HTML5: true,
	}))
	logger, _ := zap.NewDevelopment()
	e.Use(router.AccessLoggingMiddleware(logger))

	// headerの追加のため
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		ExposeHeaders: []string{"X-Showcase-User"},
	}))

	// API定義 (/api)
	api := e.Group("/api", router.TraQUserMiddleware)
	adminAPI := api.Group("", router.AdminUserMiddleware)
	groupCreatedAPI := api.Group("/groups", router.GroupCreatedUserMiddleware)
	eventCreatedAPI := api.Group("/reservations", router.EventCreatedUserMiddleware)
	{
		apiGroups := api.Group("/groups")
		adminAPIGroups := adminAPI.Group("/groups")
		{
			apiGroups.GET("", router.HandleGetGroups)
			apiGroups.POST("", router.HandlePostGroup)
			groupCreatedAPI.PATCH("/:groupid", router.HandleUpdateGroup)
			adminAPIGroups.DELETE("/:groupid", router.HandleDeleteGroup)
		}

		apiEvents := api.Group("/reservations")
		{
			apiEvents.GET("", router.HandleGetReservations)
			apiEvents.POST("", router.HandlePostReservation)
			eventCreatedAPI.PATCH("/:reservationid", router.HandleUpdateReservation)
			eventCreatedAPI.DELETE("/:reservationid", router.HandleDeleteReservation)
		}

		apiRooms := api.Group("/rooms")
		adminAPIRooms := adminAPI.Group("/rooms")
		{
			apiRooms.GET("", router.HandleGetRooms)
			adminAPIRooms.POST("", router.HandlePostRoom)
			adminAPIRooms.POST("/all", router.HandleSetRooms)
			adminAPIRooms.DELETE("/:roomid", router.HandleDeleteRoom)
		}

		apiUsers := api.Group("/users")
		{
			apiUsers.GET("", router.HandleGetUsers)
			apiUsers.GET("/me", router.HandleGetUserMe)
		}
	}

	// サーバースタート
	go func() {
		if err := e.Start(":3000"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
