// Package router is
package router

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jszwec/csvutil"
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
	DailyChannelId    string
	Origin            string
}

func (h *Handlers) SetupRoute() *echo.Echo {
	echo.NotFoundHandler = NotFoundHandler
	// echo初期化
	e := echo.New()
	e.Binder = &CustomBinder{}
	e.HTTPErrorHandler = HTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(AccessLoggingMiddleware(h.Logger))

	if len(h.SessionKey) == 0 {
		h.SessionKey = securecookie.GenerateRandomKey(32)
	}
	e.Use(session.Middleware(sessions.NewCookieStore(h.SessionKey)))

	e.Use(ServerVersionMiddleware(domain.VERSION))

	// TODO fix "portal origin"
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://portal.trap.jp", "http://localhost:8080"},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	// API定義 (/api)

	// 認証なし
	apiNoAuth := e.Group("/api")
	{
		apiNoAuth.POST("/authParams", h.HandlePostAuthParams)
		apiNoAuth.GET("/callback", h.HandleCallback)
		apiNoAuth.GET("/ical/v1/:userIDsecret", h.HandleGetiCalByPrivateID)
		apiNoAuth.GET("/version", h.HandleGetVersion)
	}

	// 認証あり (JWT認証、traQ認証)
	apiWithAuth := apiNoAuth.Group("", h.TraQUserMiddleware)
	{
		groupsAPI := apiWithAuth.Group("/groups")
		{
			groupsAPI.GET("", h.HandleGetGroups)
			groupsAPI.POST("", h.HandlePostGroup)
			groupsAPI.GET("/:groupid", h.HandleGetGroup)
			groupsAPI.PUT("/:groupid", h.HandleUpdateGroup, h.GroupAdminsMiddleware)
			groupsAPI.DELETE("/:groupid", h.HandleDeleteGroup, h.GroupAdminsMiddleware)
			groupsAPI.PUT("/:groupid/members/me", h.HandleAddMeGroup)
			groupsAPI.DELETE("/:groupid/members/me", h.HandleDeleteMeGroup)
			groupsAPI.GET("/:groupid/events", h.HandleGetEventsByGroupID)
		}

		eventsAPI := apiWithAuth.Group("/events")
		{
			eventsAPI.GET("", h.HandleGetEvents)
			eventsAPI.POST("", h.HandlePostEvent, middleware.BodyDump(h.WebhookEventHandler))
			eventsAPI.GET("/:eventid", h.HandleGetEvent)
			eventsAPI.PUT("/:eventid", h.HandleUpdateEvent, h.EventAdminsMiddleware, middleware.BodyDump(h.WebhookEventHandler))
			eventsAPI.DELETE("/:eventid", h.HandleDeleteEvent, h.EventAdminsMiddleware)
			eventsAPI.PUT("/:eventid/attendees/me", h.HandleUpsertMeEventSchedule)
			eventsAPI.POST("/:eventid/tags", h.HandleAddEventTag)
			eventsAPI.DELETE("/:eventid/tags/:tagName", h.HandleDeleteEventTag)
		}

		roomsAPI := apiWithAuth.Group("/rooms")
		{
			roomsAPI.GET("", h.HandleGetRooms)
			roomsAPI.POST("", h.HandlePostRoom)
			roomsAPI.POST("/all", h.HandleCreateVerifedRooms, h.PrevilegeUserMiddleware)
			roomsAPI.GET("/:roomid", h.HandleGetRoom)
			roomsAPI.DELETE("/:roomid", h.HandleDeleteRoom)
			roomsAPI.POST("/:roomid/verified", h.HandleVerifyRoom, h.PrevilegeUserMiddleware)
			roomsAPI.DELETE("/:roomid/verified", h.HandleUnVerifyRoom, h.PrevilegeUserMiddleware)
		}

		usersAPI := apiWithAuth.Group("/users")
		{
			usersAPI.GET("", h.HandleGetUsers)
			usersAPI.POST("/sync", h.HandleSyncUser, h.PrevilegeUserMiddleware)
			usersAPI.GET("/me", h.HandleGetUserMe)
			usersAPI.GET("/me/ical", h.HandleGetiCal)
			usersAPI.PUT("/me/ical", h.HandleUpdateiCal)
			usersAPI.GET("/me/groups", h.HandleGetMeGroupIDs)
			usersAPI.GET("/me/events", h.HandleGetMeEvents)
			usersAPI.GET("/:userid/events", h.HandleGetEventsByUserID)
			usersAPI.GET("/:userid/groups", h.HandleGetGroupIDsByUserID)
		}

		tagsAPI := apiWithAuth.Group("/tags")
		{
			tagsAPI.POST("", h.HandlePostTag)
			tagsAPI.GET("", h.HandleGetTags)
		}
	}

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/api")
		},
		Root:  "web/dist",
		HTML5: true,
	}))

	return e
}

func getConinfo(c echo.Context) *domain.ConInfo {
	info := new(domain.ConInfo)
	sess, _ := session.Get("session", c)
	str := sess.Values["userID"].(string)
	info.ReqUserID = uuid.FromStringOrNil(str)
	return info
}

type CustomBinder struct{}

func (cb *CustomBinder) Bind(i interface{}, c echo.Context) error {
	// You may use default binder
	db := new(echo.DefaultBinder)
	if err := db.Bind(i, c); err != echo.ErrUnsupportedMediaType {
		return err
	}

	// Define your custom implementation here
	rq := c.Request()
	ctype := rq.Header.Get(echo.HeaderContentType)

	if strings.HasPrefix(ctype, "text/csv") {

		buf := new(bytes.Buffer)
		_, _ = io.Copy(buf, c.Request().Body)
		data := buf.Bytes()
		if err := csvutil.Unmarshal(data, i); err != nil {
			return err
		}
		return nil
	}

	return echo.ErrUnsupportedMediaType
}
