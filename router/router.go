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

type Handlers struct {
	Repo                      repo.Repository
	InitExternalUserGroupRepo func(token string, ver repo.TraQVersion) interface {
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
			apiGroups.GET("", h.HandleGetGroups)
			apiGroups.POST("", h.HandlePostGroup)
			apiGroup := apiGroups.Group("/:groupid")
			{
				apiGroups.GET("/:groupid", h.HandleGetGroup)

				//apiGroups.PUT("/:groupid", h.HandleUpdateGroup, GroupCreatedUserMiddleware)
				apiGroups.PUT("/:groupid", h.HandleUpdateGroup)

				apiGroups.DELETE("/:groupid", h.HandleDeleteGroup, adminMiddle)

				apiGroup.PATCH("/members/me", h.HandleAddMeGroup)
				apiGroup.DELETE("/members/me", h.HandleDeleteMeGroup)
			}
		}

		apiEvents := api.Group("/events")
		{
			apiEvents.GET("", h.HandleGetEvents)
			apiEvents.POST("", h.HandlePostEvent)

			apiEvent := apiEvents.Group("/:eventid", EventIDMiddleware)
			{
				apiEvent.GET("", h.HandleGetEvent)
				apiEvent.PUT("", h.HandleUpdateEvent)
				apiEvent.DELETE("", h.HandleDeleteEvent)

				apiEvent.PATCH("/tags", h.HandleAddEventTag)
				apiEvent.DELETE("/tags/:tagName", h.HandleDeleteEventTag)
			}

		}
		apiRooms := api.Group("/rooms")
		{
			apiRooms.GET("", h.HandleGetRooms)
			apiRooms.POST("", h.HandlePostRoom, adminMiddle)
			apiRooms.POST("/private", h.HandlePostPrivateRoom)
			apiRooms.GET("/:roomid", h.HandleGetRoom)
			apiRooms.POST("/all", h.HandleSetRooms, adminMiddle)
			apiRooms.DELETE("/:roomid", h.HandleDeleteRoom, adminMiddle)
			// TODO createdBy only
			apiRooms.DELETE("/private/:roomid", h.HandleDeletePrivateRoom)
		}

		apiUsers := api.Group("/users")
		{
			apiUsers.GET("", h.HandleGetUsers)
			apiUsers.GET("/me", h.HandleGetUserMe)
		}

		apiTags := api.Group("/tags")
		{
			apiTags.POST("", h.HandlePostTag)
			apiTags.GET("", h.HandleGetTags)
		}

	}
	e.POST("/api/authParams", HandlePostAuthParams)

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:  "web/dist",
		HTML5: true,
	}))

	return e
}
