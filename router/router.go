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
	JWTKey            string
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
	apiWithAuth := apiNoAuth.Group("", h.JWTMiddleware(), h.TraQUserMiddleware)
	{
		apiGroups := apiWithAuth.Group("/groups")
		{
			apiGroups.GET("", h.HandleGetGroups)
			apiGroups.POST("", h.HandlePostGroup)
			apiGroups.GET("/:groupid", h.HandleGetGroup)
			apiGroups.PUT("/:groupid", h.HandleUpdateGroup, h.GroupAdminsMiddleware)
			apiGroups.DELETE("/:groupid", h.HandleDeleteGroup, h.GroupAdminsMiddleware)
			apiGroups.PUT("/:groupid/members/me", h.HandleAddMeGroup)
			apiGroups.DELETE("/:groupid/members/me", h.HandleDeleteMeGroup)
			apiGroups.GET("/:groupid/events", h.HandleGetEventsByGroupID)
		}

		apiEvents := apiWithAuth.Group("/events")
		{
			apiEvents.GET("", h.HandleGetEvents)
			apiEvents.POST("", h.HandlePostEvent, middleware.BodyDump(h.WebhookEventHandler))
			apiEvents.GET("/:eventid", h.HandleGetEvent)
			apiEvents.PUT("/:eventid", h.HandleUpdateEvent, h.EventAdminsMiddleware, middleware.BodyDump(h.WebhookEventHandler))
			apiEvents.DELETE("/:eventid", h.HandleDeleteEvent, h.EventAdminsMiddleware)
			apiEvents.PUT("/:eventid/attendees/me", h.HandleUpsertMeEventSchedule)
			apiEvents.POST("/:eventid/tags", h.HandleAddEventTag)
			apiEvents.DELETE("/:eventid/tags/:tagName", h.HandleDeleteEventTag)
		}

		apiRooms := apiWithAuth.Group("/rooms")
		{
			apiRooms.GET("", h.HandleGetRooms)
			apiRooms.POST("", h.HandlePostRoom)
			apiRooms.POST("/all", h.HandleCreateVerifedRooms, h.PrevilegeUserMiddleware)
			apiRooms.GET("/:roomid", h.HandleGetRoom)
			apiRooms.DELETE("/:roomid", h.HandleDeleteRoom)
			apiRooms.POST("/:roomid/verified", h.HandleVerifyRoom, h.PrevilegeUserMiddleware)
			apiRooms.DELETE("/:roomid/verified", h.HandleUnVerifyRoom, h.PrevilegeUserMiddleware)
		}

		apiUsers := apiWithAuth.Group("/users")
		{
			apiUsers.GET("", h.HandleGetUsers)
			apiUsers.POST("/sync", h.HandleSyncUser, h.PrevilegeUserMiddleware)
			apiUsers.GET("/me", h.HandleGetUserMe)
			apiUsers.GET("/me/ical", h.HandleGetiCal)
			apiUsers.PUT("/me/ical", h.HandleUpdateiCal)
			apiUsers.GET("/me/groups", h.HandleGetMeGroupIDs)
			apiUsers.GET("/me/events", h.HandleGetMeEvents)
			apiUsers.GET("/:userid/events", h.HandleGetEventsByUserID)
			apiUsers.GET("/:userid/groups", h.HandleGetGroupIDsByUserID)
		}

		apiTags := apiWithAuth.Group("/tags")
		{
			apiTags.POST("", h.HandlePostTag)
			apiTags.GET("", h.HandleGetTags)
		}

		apiWithAuth.POST("/token", h.HandleCreateToken)
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
