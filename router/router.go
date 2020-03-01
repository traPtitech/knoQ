package router

import (
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/wader/gormstore"
	"go.uber.org/zap"
)

type Handlers struct {
	Repo   // repository.Repository
	Logger *zap.Logger
}

func (h *Handlers) SetupRoute(e *echo.Group) {
	echo.NotFoundHandler = router.NotFoundHandler
	// echo初期化
	e := echo.New()
	e.HTTPErrorHandler = router.HTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	logger, _ := zap.NewDevelopment()
	e.Use(router.AccessLoggingMiddleware(logger))

	if len(SESSION_KEY) == 0 {
		SESSION_KEY = securecookie.GenerateRandomKey(32)
		fmt.Println(SESSION_KEY)
	}
	e.Use(session.Middleware(gormstore.New(db, []byte("WIP"))))
	e.Use(router.WatchCallbackMiddleware())

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
	e.POST("/api/authParams", router.HandlePostAuthParams)

}
