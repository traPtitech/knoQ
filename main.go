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
	e.Use(middleware.Static("./web/dist"))

	// headerの追加のため
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		ExposeHeaders: []string{"X-Showcase-User"},
	}))

	// API定義 (/api)
	api := e.Group("/api", myMiddleware.TraQUserMiddleware)
	api.GET("/users", router.HandleGetUsers)
	api.GET("/users/me", router.HandleGetUserMe)
	api.GET("/rooms", router.HandleGetRooms)
	api.GET("/groups", router.HandleGetGroups)
	api.POST("/groups", router.HandlePostGroup)
	api.PATCH("/groups/:groupid", router.HandleUpdateGroup)
	api.GET("/reservations", router.HandleGetReservations)
	api.POST("/reservations", router.HandlePostReservation)
	api.DELETE("/reservations/:reservationid", router.HandleDeleteReservation)
	api.PATCH("/reservations/:reservationid", router.HandleUpdateReservation)

	// 管理者専用API定義 (/api/admin)
	adminAPI := api.Group("/admin", myMiddleware.AdminUserMiddleware)
	adminAPI.POST("/rooms", router.HandlePostRoom)
	adminAPI.POST("/rooms/all", router.HandleSetRooms)
	adminAPI.DELETE("/rooms/:roomid", router.HandleDeleteRoom)
	adminAPI.DELETE("/groups/:groupid", router.HandleDeleteGroup)

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
