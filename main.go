package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"room/model"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	db, err := model.SetupDatabase()
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
	api := e.Group("/api", model.TraQUserMiddleware)
	api.GET("/hello", model.HandleGetHello) // テスト用
	api.GET("/users", model.HandleGetUsers)
	api.GET("/users/me", model.HandleGetUserMe)
	api.GET("/rooms", model.HandleGetRooms)
	api.GET("/groups", model.HandleGetGroups)
	api.POST("/groups", model.HandlePostGroup)
	api.PATCH("/groups/:groupid", model.HandleUpdateGroup)
	api.GET("/reservations", model.HandleGetReservations)
	api.POST("/reservations", model.HandlePostReservation)
	api.DELETE("/reservations/:reservationid", model.HandleDeleteReservation)
	api.PATCH("/reservations/:reservationid", model.HandleUpdateReservation)

	// 管理者専用API定義 (/api/admin)
	adminAPI := api.Group("/admin", model.AdminUserMiddleware)
	adminAPI.GET("/hello", model.HandleGetHello) // テスト用
	adminAPI.POST("/rooms", model.HandlePostRoom)
	adminAPI.POST("/rooms/all", model.HandleSetRooms)
	adminAPI.DELETE("/rooms/:roomid", model.HandleDeleteRoom)
	adminAPI.DELETE("/groups/:groupid", model.HandleDeleteGroup)

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
