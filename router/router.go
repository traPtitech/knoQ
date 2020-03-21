// Package router is
package router

import (
	"net/http"

	repo "room/repository"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wader/gormstore"
	"go.uber.org/zap"
)

type Version int64

const (
	v2 Version = iota
	v3
)

type Handlers struct {
	Repo                      repo.Repository
	InitExternalUserGroupRepo func(token string, ver Version) interface {
		repo.GroupRepository
		repo.UserRepository
	}
	ExternalRoomRepo repo.RoomRepository
	Logger           *zap.Logger
	SessionKey       []byte
}

func (h *Handlers) SetupRoute(db *gorm.DB) *echo.Echo {
	echo.NotFoundHandler = NotFoundHandler
	// echo初期化
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())

	e.Use(AccessLoggingMiddleware(h.Logger))

	e.Use(session.Middleware(gormstore.New(db, h.SessionKey)))
	e.Use(WatchCallbackMiddleware())

	// TODO fix "portal origin"
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://portal.trap.jp", "http://localhost:8080"},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	// API定義 (/api)
	api := e.Group("/api", TraQUserMiddleware)
	{
		adminMiddle := AdminUserMiddleware

		apiGroups := api.Group("/groups")
		{
			apiGroups.GET("", HandleGetGroups)
			apiGroups.POST("", h.HandlePostGroup)
			apiGroup := apiGroups.Group("/:groupid")
			{
				apiGroups.GET("/:groupid", h.HandleGetGroup)
				apiGroups.PUT("/:groupid", HandleUpdateGroup, GroupCreatedUserMiddleware)
				apiGroups.DELETE("/:groupid", HandleDeleteGroup, adminMiddle)

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
			apiUsers.GET("", h.HandleGetUsers)
			apiUsers.GET("/me", h.HandleGetUserMe)
		}

		apiTags := api.Group("/tags")
		{
			apiTags.POST("", HandlePostTag)
			apiTags.GET("", HandleGetTags)
		}

	}
	e.POST("/api/authParams", HandlePostAuthParams)

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "web/dist",
		HTML5: true,
	}))

	return e
}
