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
			groupsAPI.PUT("/:groupid/members/me", h.HandleAddMeGroup)
			groupsAPI.DELETE("/:groupid/members/me", h.HandleDeleteMeGroup)
			groupsAPI.GET("/:groupid/events", h.HandleGetEventsByGroupID)

			// グループ管理者権限が必要
			groupsAPIWithAdminAuth := groupsAPI.Group("", h.GroupAdminsMiddleware)
			{
				groupsAPIWithAdminAuth.PUT("/:groupid/members/:userid", h.HandleUpdateGroup)
				groupsAPIWithAdminAuth.DELETE("/:groupid/members/:userid", h.HandleDeleteGroup)
			}
		}

		eventsAPI := apiWithAuth.Group("/events")
		{
			eventsAPI.GET("", h.HandleGetEvents)
			eventsAPI.POST("", h.HandlePostEvent, middleware.BodyDump(h.WebhookEventHandler))
			eventsAPI.GET("/:eventid", h.HandleGetEvent)
			eventsAPI.PUT("/:eventid/attendees/me", h.HandleUpsertMeEventSchedule)
			eventsAPI.POST("/:eventid/tags", h.HandleAddEventTag)
			eventsAPI.DELETE("/:eventid/tags/:tagName", h.HandleDeleteEventTag)

			// イベント管理者権限が必要
			eventsAPIWithAdminAuth := eventsAPI.Group("", h.EventAdminsMiddleware)
			{
				eventsAPIWithAdminAuth.PUT("/:eventid", h.HandleUpdateEvent, middleware.BodyDump(h.WebhookEventHandler))
				eventsAPIWithAdminAuth.DELETE("/:eventid", h.HandleDeleteEvent)
			}
		}

		roomsAPI := apiWithAuth.Group("/rooms")
		{
			roomsAPI.GET("", h.HandleGetRooms)
			roomsAPI.POST("", h.HandlePostRoom)
			roomsAPI.GET("/:roomid", h.HandleGetRoom)
			roomsAPI.DELETE("/:roomid", h.HandleDeleteRoom)

			// サービス管理者権限が必要
			roomsAPIWithPrevilegeAuth := roomsAPI.Group("", h.PrevilegeUserMiddleware)
			{
				roomsAPIWithPrevilegeAuth.POST("/all", h.HandleCreateVerifedRooms)
				roomsAPIWithPrevilegeAuth.POST("/:roomid/verified", h.HandleVerifyRoom)
				roomsAPIWithPrevilegeAuth.DELETE("/:roomid/verified", h.HandleUnVerifyRoom)
			}
		}

		usersAPI := apiWithAuth.Group("/users")
		{
			usersAPI.GET("", h.HandleGetUsers)
			usersAPI.GET("/me", h.HandleGetUserMe)
			usersAPI.GET("/me/ical", h.HandleGetiCal)
			usersAPI.PUT("/me/ical", h.HandleUpdateiCal)
			usersAPI.GET("/me/groups", h.HandleGetMeGroupIDs)
			usersAPI.GET("/me/events", h.HandleGetMeEvents)
			usersAPI.GET("/:userid/events", h.HandleGetEventsByUserID)
			usersAPI.GET("/:userid/groups", h.HandleGetGroupIDsByUserID)

			// サービス管理者権限が必要
			usersAPIWithPrevilegeAuth := usersAPI.Group("", h.PrevilegeUserMiddleware)
			{
				usersAPIWithPrevilegeAuth.PATCH("/:userid/privileged", h.HandleGrantPrivilege)
				usersAPIWithPrevilegeAuth.POST("/sync", h.HandleSyncUser)
			}
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
