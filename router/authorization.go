package router

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	traQrandom "github.com/traPtitech/traQ/utils/random"
)

var verifierCache = cache.New(5*time.Minute, 10*time.Minute)
var stateCache = cache.New(5*time.Minute, 10*time.Minute)

type AuthParams struct {
	URL string `json:"url"`
}

func (h *Handlers) HandlePostAuthParams(c echo.Context) error {
	url, state, codeVerifier := h.Repo.GetOAuthURL()

	// cache codeVerifier
	sess, err := session.Get("session", c)
	if err != nil {
		setMaxAgeMinus(c)
		return unauthorized(err, needAuthorization(true),
			message("please try again"))
	}

	sessionID, ok := sess.Values["ID"].(string)
	if !ok {
		sessionID = traQrandom.SecureAlphaNumeric(10)
		sess.Values["ID"] = sessionID
		sess.Options = &h.SessionOption
		_ = sess.Save(c.Request(), c.Response())
	}
	// cache
	verifierCache.Set(sessionID, codeVerifier, cache.DefaultExpiration)
	stateCache.Set(sessionID, state, cache.DefaultExpiration)

	authParams := &AuthParams{
		URL: url,
	}

	return c.JSON(http.StatusCreated, authParams)
}

func (h *Handlers) HandleCallback(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		setMaxAgeMinus(c)
		return unauthorized(err, needAuthorization(true),
			message("please try again"))
	}
	sessionID, ok := sess.Values["ID"].(string)
	if !ok {
		return internalServerError(errors.New("session error"))
	}
	codeVerifier, ok := verifierCache.Get(sessionID)
	if !ok {
		return internalServerError(errors.New("codeVerifier is not cached"))
	}
	state, ok := stateCache.Get(sessionID)
	if !ok {
		return internalServerError(errors.New("state is not cached"))
	}
	user, err := h.Repo.LoginUser(c.QueryString(), state.(string), codeVerifier.(string))
	if err != nil {
		return internalServerError(err)
	}

	sess.Values["userID"] = user.ID.String()
	sess.Options = &h.SessionOption
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return internalServerError(err)
	}
	return c.Redirect(http.StatusFound, "/callback")
}
