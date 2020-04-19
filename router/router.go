// Package router is
package router

import (
	"net/http"
	"strings"

	"room/router/service"

	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wader/gormstore"
	"go.uber.org/zap"
)

type Handlers struct {
	service.Dao
	Logger        *zap.Logger
	SessionKey    []byte
	SessionOption sessions.Options
	ClientID      string
	WebhookID     string
	WebhookSecret string
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
	e.Use(h.WatchCallbackMiddleware())

	// TODO fix "portal origin"
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://portal.trap.jp", "http://localhost:8080"},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	// API定義 (/api)
	api := e.Group("/api", h.TraQUserMiddleware)
	{
		adminMiddle := h.AdminUserMiddleware

		apiGroups := api.Group("/groups")
		{
			apiGroups.GET("", h.HandleGetGroups)
			apiGroups.POST("", h.HandlePostGroup)
			apiGroup := apiGroups.Group("/:groupid")
			{
				apiGroup.GET("", h.HandleGetGroup)

				apiGroup.PUT("", h.HandleUpdateGroup, h.GroupCreatedUserMiddleware)
				apiGroup.DELETE("", h.HandleDeleteGroup, h.GroupCreatedUserMiddleware)

				apiGroup.PUT("/members/me", h.HandleAddMeGroup)
				apiGroup.DELETE("/members/me", h.HandleDeleteMeGroup)

				apiGroup.GET("/events", h.HandleGetEventsByGroupID)
			}
		}

		apiEvents := api.Group("/events")
		{
			apiEvents.GET("", h.HandleGetEvents)
			apiEvents.POST("", h.HandlePostEvent, middleware.BodyDump(h.WebhookEventHandler))

			apiEvent := apiEvents.Group("/:eventid")
			{
				apiEvent.GET("", h.HandleGetEvent)
				apiEvent.PUT("", h.HandleUpdateEvent, h.EventCreatedUserMiddleware, middleware.BodyDump(h.WebhookEventHandler))
				apiEvent.DELETE("", h.HandleDeleteEvent, h.EventCreatedUserMiddleware)

				apiEvent.POST("/tags", h.HandleAddEventTag)
				apiEvent.DELETE("/tags/:tagName", h.HandleDeleteEventTag)
			}

		}
		apiRooms := api.Group("/rooms")
		{
			apiRooms.GET("", h.HandleGetRooms)
			apiRooms.POST("", h.HandlePostRoom, adminMiddle)
			apiRooms.POST("/all", h.HandleSetRooms, adminMiddle)

			apiRooms.POST("/private", h.HandlePostPrivateRoom)

			apiRoom := apiRooms.Group("/:roomid")
			{
				apiRoom.GET("", h.HandleGetRoom)
				apiRoom.DELETE("", h.HandleDeleteRoom, adminMiddle)
				apiRoom.GET("/events", h.HandleGetEventsByRoomID)
			}
			apiRooms.DELETE("/private/:roomid", h.HandleDeletePrivateRoom, h.RoomCreatedUserMiddleware)
		}

		apiUsers := api.Group("/users")
		{
			apiUsers.GET("", h.HandleGetUsers)

			apiUsers.GET("/me", h.HandleGetUserMe)
			apiUsers.GET("/me/groups", h.HandleGetMeGroupIDs)
			apiUsers.GET("/me/events", h.HandleGetMeEvents)

			apiUser := apiUsers.Group("/:userid")
			{
				apiUser.GET("/events", h.HandleGetEventsByUserID)
				apiUser.GET("/groups", h.HandleGetGroupIDsByUserID)
			}
		}

		apiTags := api.Group("/tags")
		{
			apiTags.POST("", h.HandlePostTag)
			apiTags.GET("", h.HandleGetTags)
		}

		apiActivity := api.Group("/activity")
		{
			apiActivity.GET("/events", h.HandleGetEventActivities)
		}

	}
	e.POST("/api/authParams", h.HandlePostAuthParams)

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/api")
		},
		Root:  "web/dist",
		HTML5: true,
	}))

	return e
}
