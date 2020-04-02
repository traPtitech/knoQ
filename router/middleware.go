package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	log "room/logging"
	repo "room/repository"
	"room/utils"
	"strconv"
	"strings"
	"time"

	traQutils "github.com/traPtitech/traQ/utils"

	"github.com/gofrs/uuid"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const requestUserStr string = "Request-User"
const authScheme string = "Bearer"

var traQjson = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "traq",
}

type OauthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
}

type UserID struct {
	Value uuid.UUID `json:"userId"`
}

// CreatedByGetter get created user
type CreatedByGetter interface {
	GetCreatedBy() (uuid.UUID, error)
}

func AccessLoggingMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			req := c.Request()
			res := c.Response()
			tmp := &log.HTTPPayload{
				RequestMethod: req.Method,
				Status:        res.Status,
				UserAgent:     req.UserAgent(),
				RemoteIP:      c.RealIP(),
				Referer:       req.Referer(),
				Protocol:      req.Proto,
				RequestURL:    req.URL.String(),
				RequestSize:   req.Header.Get(echo.HeaderContentLength),
				ResponseSize:  strconv.FormatInt(res.Size, 10),
				Latency:       strconv.FormatFloat(stop.Sub(start).Seconds(), 'f', 9, 64) + "s",
			}
			httpCode := res.Status
			switch {
			case httpCode >= 500:
				errorRuntime, ok := c.Get("Error-Runtime").(error)
				if ok {
					tmp.ErrorLocation = errorRuntime.Error()
				} else {
					tmp.ErrorLocation = "no data"
				}
				logger.Info("server error", zap.Object("field", tmp))
			case httpCode >= 400:
				errorRuntime, ok := c.Get("Error-Runtime").(error)
				if ok {
					tmp.ErrorLocation = errorRuntime.Error()
				} else {
					tmp.ErrorLocation = "no data"
				}
				logger.Info("client error", zap.Object("field", tmp))
			case httpCode >= 300:
				logger.Info("redirect", zap.Object("field", tmp))
			case httpCode >= 200:
				logger.Info("success", zap.Object("field", tmp))
			}
			return nil
		}
	}
}

// WatchCallbackMiddleware /callback?code= を監視
func WatchCallbackMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if path != "/callback" {
				return next(c)
			}
			code := c.QueryParam("code")

			sess, _ := session.Get("session", c)
			sessionID, ok := sess.Values["ID"].(string)
			if !ok {
				fmt.Println("err")
				return internalServerError()
			}
			codeVerifier, ok := verifierCache.Get(sessionID)
			if !ok {
				return internalServerError()
			}

			form := url.Values{}
			form.Add("grant_type", "authorization_code")
			form.Add("client_id", "1iZopJ2qP63BaJYkQxhlVzCdrG8h1tDHMXm7")
			form.Add("code", code)
			form.Add("code_verifier", codeVerifier.(string))

			body := strings.NewReader(form.Encode())

			req, err := http.NewRequest("POST", "https://q.trap.jp/api/1.0/oauth2/token", body)
			if err != nil {
				return next(c)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return internalServerError()
			}
			if res.StatusCode >= 300 {
				return internalServerError()
			}

			data, _ := ioutil.ReadAll(res.Body)
			oauthRes := new(OauthResponse)
			json.Unmarshal(data, oauthRes)
			token := oauthRes.AccessToken

			bytes, _ := utils.GetUserMe(token)
			userID := new(UserID)
			json.Unmarshal(bytes, userID)

			sess.Values["authorization"] = token
			sess.Values["userID"] = userID.Value.String()
			sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			err = sess.Save(c.Request(), c.Response())
			if err != nil {
				return internalServerError()
			}

			return next(c)
		}
	}
}

// TraQUserMiddleware traQユーザーか判定するミドルウェア
func (h *Handlers) TraQUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return unauthorized()
		}
		auth, ok := sess.Values["authorization"].(string)
		if !ok {
			sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			sess.Values["ID"] = traQutils.RandAlphabetAndNumberString(10)
			sess.Save(c.Request(), c.Response())
			return unauthorized()
		}
		if auth == "" {
			return unauthorized()
		}
		// TODO get admin from db
		setRequestUserIsAdmin(c, h.Repo)
		return next(c)
	}
}

// AdminUserMiddleware 管理者ユーザーか判定するミドルウェア
func (h *Handlers) AdminUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		isAdmin := getRequestUserIsAdmin(c)

		// 判定
		if !isAdmin {
			return forbidden(
				message("You are not admin user."),
				specification("Only admin user can request."),
			)
		}

		return next(c)
	}
}

// GroupCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func GroupCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUserID, _ := getRequestUserID(c)
		g := new(repo.Group)
		var err error
		g.ID, err = getRequestGroupID(c)
		if err != nil || g.ID == uuid.Nil {
			return notFound()
		}
		IsVerigy, err := verifyCreatedUser(g, requestUserID)
		if err != nil {
			return internalServerError()
		}
		if !IsVerigy {
			return badRequest(
				message("You are not user by whom this group is created."),
				specification("Only the author can request."),
			)
		}
		return next(c)
	}
}

// EventCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func EventCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUserID, _ := getRequestUserID(c)
		event := new(repo.Event)
		var err error
		event.ID, err = getRequestEventID(c)
		if err != nil {
			return internalServerError()
		}

		IsVerigy, err := verifyCreatedUser(event, requestUserID)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
			}
			return internalServerError()
		}
		if !IsVerigy {
			return badRequest(
				message("You are not user by whom this even is created."),
				specification("Only the author can request."),
			)
		}

		return next(c)
	}
}

func EventIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		event := new(repo.Event)
		var err error
		event.ID, err = uuid.FromString(c.Param("eventid"))
		if err != nil || event.ID == uuid.Nil {
			return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
		}
		if err := repo.DB.Select("id").First(&event).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return notFound(message(fmt.Sprintf("EventID: %v does not exist.", c.Param("eventid"))))
			}
			return internalServerError()
		}
		c.Set("EventID", event.ID)

		return next(c)
	}
}

func GroupIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		group := new(repo.Group)
		var err error
		group.ID, err = uuid.FromString(c.Param("groupid"))
		if err != nil || group.ID == uuid.Nil {
			return notFound(message(fmt.Sprintf("GroupID: %v does not exist.", c.Param("groupid"))))
		}
		if err := repo.DB.Select("id").First(&group).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return notFound(message(fmt.Sprintf("GroupID: %v does not exist.", c.Param("groupid"))))
			}
			return internalServerError()
		}
		c.Set("GroupID", group.ID)

		return next(c)

	}
}

// verifyCreatedUser verify that request-user and created-user are the same
func verifyCreatedUser(cbg CreatedByGetter, requestUser uuid.UUID) (bool, error) {
	createdByUser, err := cbg.GetCreatedBy()
	if err != nil {
		return false, err
	}
	if createdByUser != requestUser {
		return false, nil
	}
	return true, nil
}

func getRequestUserID(c echo.Context) (uuid.UUID, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.FromString(sess.Values["userID"].(string))
}

func setRequestUserIsAdmin(c echo.Context, repo repo.UserRepository) error {
	userID, _ := getRequestUserID(c)
	user, err := repo.GetUser(userID)
	if err != nil {
		return err
	}
	c.Set("IsAdmin", user.Admin)
	return nil
}

func getRequestUserIsAdmin(c echo.Context) bool {
	return c.Get("IsAdmin").(bool)
}

func getRequestUserToken(c echo.Context) (string, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return "", err
	}
	token, ok := sess.Values["authorization"].(string)
	if !ok {
		return "", errors.New("error")
	}

	return token, nil
}

func deleteRequestUserToken(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Values["authorization"] = ""
	err = sess.Save(c.Request(), c.Response())
	return err
}

// getRequestEventID :eventidを返します
func getRequestEventID(c echo.Context) (uuid.UUID, error) {

	eventID, err := uuid.FromString(c.Param("eventid"))
	if err != nil {
		return uuid.Nil, errors.New("EventID is not uuid")
	}
	return eventID, nil
}

// getRequestGroupID :groupidを返します
func getRequestGroupID(c echo.Context) (uuid.UUID, error) {
	groupID, err := uuid.FromString(c.Param("groupid"))
	if err != nil {
		return uuid.Nil, errors.New("GroupID is not uuid")
	}
	return groupID, nil
}
