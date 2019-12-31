package main

import (
	"context"
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

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
)

func main() {
	db, err := repo.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	echo.NotFoundHandler = router.NotFoundHandler
	// echo初期化
	e := echo.New()
	e.HTTPErrorHandler = router.HTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "web/dist",
		HTML5: true,
	}))
	logger, _ := zap.NewDevelopment()
	e.Use(router.AccessLoggingMiddleware(logger))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	// headerの追加のため
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		ExposeHeaders: []string{"X-Showcase-User"},
	}))

	// API定義 (/api)
	api := e.Group("/api", router.TraQUserMiddleware)
	{
		adminMiddle := router.AdminUserMiddleware

		apiGroups := api.Group("/groups")
		{
			apiGroups.GET("", router.HandleGetGroups)
			apiGroups.POST("", router.HandlePostGroup)
			apiGroup := apiGroups.Group("/:groupid", router.GroupIDMiddleware)
			{
				apiGroup.GET("", router.HandleGetGroup)
				apiGroup.PUT("", router.HandleUpdateGroup, router.GroupCreatedUserMiddleware)
				apiGroup.DELETE("", router.HandleDeleteGroup, adminMiddle)

				// apiGroup.PATCH("/tags", router.HandleAddGroupTag)
				// apiGroup.DELETE("/tags/:tagid", router.HandleDeleteGroupTag)

				apiGroup.PATCH("/members/me", router.HandleAddMeGroup)
				apiGroup.DELETE("/members/me", router.HandleDeleteMeGroup)
			}
		}

		apiEvents := api.Group("/events")
		{
			apiEvents.GET("", router.HandleGetEvents)
			apiEvents.POST("", router.HandlePostEvent)

			apiEvent := apiEvents.Group("/:eventid", router.EventIDMiddleware)
			{
				apiEvent.GET("", router.HandleGetEvent)
				apiEvent.PUT("", router.HandleUpdateEvent, router.EventCreatedUserMiddleware)
				apiEvent.DELETE("", router.HandleDeleteEvent, router.EventCreatedUserMiddleware)

				apiEvent.PATCH("/tags", router.HandleAddEventTag)
				apiEvent.DELETE("/tags/:tagid", router.HandleDeleteEventTag)
			}

		}
		apiRooms := api.Group("/rooms")
		{
			apiRooms.GET("", router.HandleGetRooms)
			apiRooms.POST("", router.HandlePostRoom, adminMiddle)
			apiRooms.GET("/:roomid", router.HandleGetRoom)
			apiRooms.POST("/all", router.HandleSetRooms, adminMiddle)
			apiRooms.DELETE("/:roomid", router.HandleDeleteRoom, adminMiddle)
		}

		apiUsers := api.Group("/users")
		{
			apiUsers.GET("", router.HandleGetUsers)
			apiUsers.GET("/me", router.HandleGetUserMe)
		}

		apiTags := api.Group("/tags")
		{
			apiTags.POST("", router.HandlePostTag)
			apiTags.GET("", router.HandleGetTags)
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
