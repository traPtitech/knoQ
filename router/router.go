// Package router is
package router

import (
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wader/gormstore"
	"go.uber.org/zap"
)

type Handlers struct {
	Repo   interface{} // TODO fix
	Logger *zap.Logger
}

func SetupRoute(SESSION_KEY []byte, db *gorm.DB) {
	echo.NotFoundHandler = NotFoundHandler
	// echo初期化
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	logger, _ := zap.NewDevelopment()
	e.Use(AccessLoggingMiddleware(logger))

	if len(SESSION_KEY) == 0 {
		SESSION_KEY = securecookie.GenerateRandomKey(32)
		fmt.Println(SESSION_KEY)
	}
	e.Use(session.Middleware(gormstore.New(db, []byte("WIP"))))
	e.Use(WatchCallbackMiddleware())

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "web/dist",
		HTML5: true,
	}))

	// TODO fix "portal origin"
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://portal.trap.jp", "localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	// API定義 (/api)
	api := e.Group("/api", TraQUserMiddleware)
	{
		adminMiddle := AdminUserMiddleware

		apiGroups := api.Group("/groups")
		{
			apiGroups.GET("", HandleGetGroups)
			apiGroups.POST("", HandlePostGroup)
			apiGroup := apiGroups.Group("/:groupid", GroupIDMiddleware)
			{
				apiGroup.GET("", HandleGetGroup)
				apiGroup.PUT("", HandleUpdateGroup, GroupCreatedUserMiddleware)
				apiGroup.DELETE("", HandleDeleteGroup, adminMiddle)

				// apiGroup.PATCH("/tags", HandleAddGroupTag)
				// apiGroup.DELETE("/tags/:tagid", HandleDeleteGroupTag)

				apiGroup.PATCH("/members/me", HandleAddMeGroup)
				apiGroup.DELETE("/members/me", HandleDeleteMeGroup)
			}
		}

		apiEvents := api.Group("/events")
		{
			apiEvents.GET("", HandleGetEvents)
			apiEvents.POST("", HandlePostEvent)

			apiEvent := apiEvents.Group("/:eventid", EventIDMiddleware)
			{
				apiEvent.GET("", HandleGetEvent)
				apiEvent.PUT("", HandleUpdateEvent, EventCreatedUserMiddleware)
				apiEvent.DELETE("", HandleDeleteEvent, EventCreatedUserMiddleware)

				apiEvent.PATCH("/tags", HandleAddEventTag)
				apiEvent.DELETE("/tags/:tagid", HandleDeleteEventTag)
			}

		}
		apiRooms := api.Group("/rooms")
		{
			apiRooms.GET("", HandleGetRooms)
			apiRooms.POST("", HandlePostRoom, adminMiddle)
			apiRooms.GET("/:roomid", HandleGetRoom)
			apiRooms.POST("/all", HandleSetRooms, adminMiddle)
			apiRooms.DELETE("/:roomid", HandleDeleteRoom, adminMiddle)
		}

		apiUsers := api.Group("/users")
		{
			apiUsers.GET("", HandleGetUsers)
			apiUsers.GET("/me", HandleGetUserMe)
		}

		apiTags := api.Group("/tags")
		{
			apiTags.POST("", HandlePostTag)
			apiTags.GET("", HandleGetTags)
		}

	}
	e.POST("/api/authParams", HandlePostAuthParams)

}
