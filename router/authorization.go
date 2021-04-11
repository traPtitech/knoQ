package router

import (
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
	url, state, codeVerifier := h.repo.GetOAuthURL()

	// cache codeVerifier
	sess, err := session.Get("session", c)
	if err != nil {
		return internalServerError(err)
	}

	sessionID, ok := sess.Values["ID"].(string)
	if !ok {
		sessionID = traQrandom.SecureAlphaNumeric(10)

		sess.Values["ID"] = sessionID
		sess.Options = &h.SessionOption
		sess.Save(c.Request(), c.Response())
	}
	// cache
	verifierCache.Set(sessionID, codeVerifier, cache.DefaultExpiration)
	stateCache.Set(sessionID, state, cache.DefaultExpiration)

	authParams := &AuthParams{
		URL: url,
	}

	return c.JSON(http.StatusCreated, authParams)
}
