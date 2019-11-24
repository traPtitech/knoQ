package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	myMiddleware "room/middleware"
	repo "room/repository"
	"room/router"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	db, err := repo.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// echo初期化
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Secure())
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "web/dist",
		HTML5: true,
	}))

	// headerの追加のため
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		ExposeHeaders: []string{"X-Showcase-User"},
	}))

	// API定義 (/api)
	api := e.Group("/api", myMiddleware.TraQUserMiddleware)
	adminAPI := api.Group("", myMiddleware.AdminUserMiddleware)
	groupCreatedAPI := api.Group("/groups", myMiddleware.GroupCreatedUserMiddleware)
	eventCreatedAPI := api.Group("/reservations", myMiddleware.EventCreatedUserMiddleware)
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
