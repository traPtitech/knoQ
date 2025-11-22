package router

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/traPtitech/knoQ/utils/random"
)

var (
	verifierCache = cache.New(5*time.Minute, 10*time.Minute)
	stateCache    = cache.New(5*time.Minute, 10*time.Minute)
)

type AuthParams struct {
	URL string `json:"url"`
}

func (h *Handlers) HandlePostAuthParams(c echo.Context) error {
	ctx := c.Request().Context()
	url, state, codeVerifier := h.Service.GetOAuthURL(ctx)

	// cache codeVerifier
	sess, err := session.Get("session", c)
	if err != nil {
		setMaxAgeMinus(c)
		return unauthorized(err, needAuthorization(true),
			message("please try again"))
	}

	sessionID, ok := sess.Values["ID"].(string)
	if !ok {
		sessionID = random.AlphaNumeric(10, true)
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

	ctx := c.Request().Context()
	user, err := h.Service.LoginUser(ctx, c.QueryString(), state.(string), codeVerifier.(string))
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
