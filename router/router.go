// Package router is
package router

import (
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Handlers struct {
	Repo              domain.Repository
	Logger            *zap.Logger
	SessionKey        []byte
	SessionOption     sessions.Options
	ClientID          string
	WebhookID         string
	WebhookSecret     string
	ActivityChannelID string
	Origin            string
}

func (h *Handlers) SetupRoute() *echo.Echo {
	echo.NotFoundHandler = NotFoundHandler
	// echo初期化
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(AccessLoggingMiddleware(h.Logger))

	if len(h.SessionKey) == 0 {
		h.SessionKey = securecookie.GenerateRandomKey(32)
	}
	e.Use(session.Middleware(sessions.NewCookieStore(h.SessionKey)))

	// TODO fix "portal origin"
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://portal.trap.jp", "http://localhost:8080"},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	// API定義 (/api)
	api := e.Group("/api", h.TraQUserMiddleware)
	{
		previlegeMiddle := h.PrevilegeUserMiddleware

		apiGroups := api.Group("/groups")
		{
			apiGroups.GET("", h.HandleGetGroups)
			apiGroups.POST("", h.HandlePostGroup)
			apiGroup := apiGroups.Group("/:groupid")
			{
				apiGroup.GET("", h.HandleGetGroup)

				apiGroup.PUT("", h.HandleUpdateGroup, h.GroupAdminsMiddleware)
				apiGroup.DELETE("", h.HandleDeleteGroup, h.GroupAdminsMiddleware)

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
				apiEvent.PUT("", h.HandleUpdateEvent, h.EventAdminsMiddleware, middleware.BodyDump(h.WebhookEventHandler))
				apiEvent.DELETE("", h.HandleDeleteEvent, h.EventAdminsMiddleware)

				apiEvent.POST("/tags", h.HandleAddEventTag)
				apiEvent.DELETE("/tags/:tagName", h.HandleDeleteEventTag)
			}

		}
		apiRooms := api.Group("/rooms")
		{
			apiRooms.GET("", h.HandleGetRooms)
			apiRooms.POST("", h.HandlePostRoom)
			apiRooms.POST("/all", h.HandleCreateVerifedRooms, previlegeMiddle)

			apiRoom := apiRooms.Group("/:roomid")
			{
				apiRoom.GET("", h.HandleGetRoom)
				apiRoom.DELETE("", h.HandleDeleteRoom)

				apiRooms.POST("/verified", h.HandleVerifyRoom, previlegeMiddle)
				apiRooms.DELETE("/verified", h.HandleUnVerifyRoom, previlegeMiddle)
			}
		}

		apiUsers := api.Group("/users")
		{
			apiUsers.GET("", h.HandleGetUsers)
			apiUsers.POST("/sync", h.HandleSyncUser, previlegeMiddle)

			apiUsers.GET("/me", h.HandleGetUserMe)
			apiUsers.GET("/me/ical", h.HandleGetiCal)
			apiUsers.PUT("/me/ical", h.HandleUpdateiCal)
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

		// apiActivity := api.Group("/activity")
		// {
		// apiActivity.GET("/events", h.HandleGetEventActivities)
		// }

	}
	e.POST("/api/authParams", h.HandlePostAuthParams)
	// TODO
	e.GET("/api/callback", nil)
	e.GET("/api/ical/v1/:userIDsecret", h.HandleGetiCalByPrivateID)

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/api")
		},
		Root:  "web/dist",
		HTML5: true,
	}))

	return e
}

func getConinfo(c echo.Context) (info *domain.ConInfo) {
	sess, _ := session.Get("session", c)
	info.ReqUserID = sess.Values["userID"].(uuid.UUID)
	return info
}
