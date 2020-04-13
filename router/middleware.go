package router

import (
	"encoding/json"
	"errors"
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
				errorRuntime, ok := c.Get("Error").(error)
				if ok {
					tmp.Error = errorRuntime.Error()
				} else {
					tmp.Error = "no data"
				}
				logger.Info("server error", zap.Object("field", tmp))
			case httpCode >= 400:
				errorRuntime, ok := c.Get("Error").(error)
				if ok {
					tmp.Error = errorRuntime.Error()
				} else {
					tmp.Error = "no data"
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
func (h *Handlers) WatchCallbackMiddleware() echo.MiddlewareFunc {
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
				return internalServerError(errors.New("sessionID can not parse string"))
			}
			codeVerifier, ok := verifierCache.Get(sessionID)
			if !ok {
				return internalServerError(errors.New("codeVerifier is not cached"))
			}

			token, err := requestOAuth(h.ClientID, code, codeVerifier.(string))
			if err != nil {
				return internalServerError(err)
			}

			// TODO fix
			bytes, _ := utils.GetUserMe(token)
			userID := new(UserID)
			json.Unmarshal(bytes, userID)

			sess.Values["authorization"] = token
			sess.Values["userID"] = userID.Value.String()
			// sess.Options = &h.SessionOption
			err = sess.Save(c.Request(), c.Response())
			if err != nil {
				return internalServerError(err)
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
			return unauthorized(err)
		}
		auth, ok := sess.Values["authorization"].(string)
		if !ok {
			sess.Options = &h.SessionOption
			sess.Values["ID"] = traQutils.RandAlphabetAndNumberString(10)
			sess.Save(c.Request(), c.Response())
			return unauthorized(err)
		}
		if auth == "" {
			return unauthorized(err)
		}
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
				errors.New("not admin"),
				message("You are not admin user."),
				specification("Only admin user can request."),
			)
		}

		return next(c)
	}
}

// GroupCreatedUserMiddleware グループ作成ユーザーか判定するミドルウェア
func (h *Handlers) GroupCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUserID, _ := getRequestUserID(c)
		token, _ := getRequestUserToken(c)
		groupID, err := getPathGroupID(c)
		if err != nil {
			return notFound(err)
		}
		group, err := h.Dao.GetGroup(token, groupID)
		if err != nil {
			return judgeErrorResponse(err)
		}
		if group.CreatedBy != requestUserID || group.IsTraQGroup {
			return forbidden(
				errors.New("not createdBy"),
				message("You are not user by whom this group is created."),
				specification("Only the author can request."),
			)
		}
		return next(c)
	}
}

// EventCreatedUserMiddleware イベント作成ユーザーか判定するミドルウェア
func (h *Handlers) EventCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUserID, _ := getRequestUserID(c)
		eventID, err := getPathEventID(c)
		if err != nil {
			return notFound(err)
		}
		event, err := h.Repo.GetEvent(eventID)
		if err != nil {
			return judgeErrorResponse(err)
		}
		if event.CreatedBy != requestUserID {
			return forbidden(
				errors.New("not createdBy"),
				message("You are not user by whom this even is created."),
				specification("Only the author can request."),
			)
		}

		return next(c)
	}
}

// RoomCreatedUserMiddleware イベント作成ユーザーか判定するミドルウェア
func (h *Handlers) RoomCreatedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUserID, _ := getRequestUserID(c)
		roomID, err := getPathRoomID(c)
		if err != nil {
			return notFound(err)
		}
		room, err := h.Repo.GetRoom(roomID)
		if err != nil {
			return judgeErrorResponse(err)
		}
		if room.CreatedBy != requestUserID {
			return forbidden(
				errors.New("not createdBy"),
				message("You are not user by whom this even is created."),
				specification("Only the author can request."),
			)
		}

		return next(c)
	}
}

func requestOAuth(clientID, code, codeVerifier string) (token string, err error) {
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", clientID)
	form.Add("code", code)
	form.Add("code_verifier", codeVerifier)

	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", "https://q.trap.jp/api/1.0/oauth2/token", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 300 {
		return "", err
	}

	data, _ := ioutil.ReadAll(res.Body)
	oauthRes := new(OauthResponse)
	json.Unmarshal(data, oauthRes)

	token = oauthRes.AccessToken
	return
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

// getPathEventID :eventidを返します
func getPathEventID(c echo.Context) (uuid.UUID, error) {

	eventID, err := uuid.FromString(c.Param("eventid"))
	if err != nil {
		return uuid.Nil, errors.New("EventID is not uuid")
	}
	return eventID, nil
}

// getPathGroupID :groupidを返します
func getPathGroupID(c echo.Context) (uuid.UUID, error) {
	groupID, err := uuid.FromString(c.Param("groupid"))
	if err != nil {
		return uuid.Nil, errors.New("GroupID is not uuid")
	}
	return groupID, nil
}

// getPathRoomID :roomidを返します
func getPathRoomID(c echo.Context) (uuid.UUID, error) {
	roomID, err := uuid.FromString(c.Param("roomid"))
	if err != nil {
		return uuid.Nil, errors.New("RoomID is not uuid")
	}
	return roomID, nil
}

// getPathUserID :useridを返します
func getPathUserID(c echo.Context) (uuid.UUID, error) {
	userID, err := uuid.FromString(c.Param("userid"))
	if err != nil {
		return uuid.Nil, errors.New("UserID is not uuid")
	}
	return userID, nil
}
