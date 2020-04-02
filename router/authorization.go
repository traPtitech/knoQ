package router

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	traQutils "github.com/traPtitech/traQ/utils"
)

var verifierCache = cache.New(5*time.Minute, 10*time.Minute)

type AuthParams struct {
	ClientID      string `json:"clientId"`
	State         string `json:"state"`
	CodeChallenge string `json:"codeChallenge"`
}

func (h *Handlers) HandlePostAuthParams(c echo.Context) error {
	codeVerifier := traQutils.RandAlphabetAndNumberString(43)

	// cache codeVerifier
	sess, err := session.Get("session", c)
	if err != nil {
		return internalServerError()
	}
	// sess.Values["ID"] = traQutils.RandAlphabetAndNumberString(10)
	// sess.Save(c.Request(), c.Response())
	sessionID, ok := sess.Values["ID"].(string)
	if !ok {
		sess.Options = &h.SessionOption
		sessionID = traQutils.RandAlphabetAndNumberString(10)
		sess.Values["ID"] = sessionID
		sess.Save(c.Request(), c.Response())
	}
	verifierCache.Set(sessionID, codeVerifier, cache.DefaultExpiration)
	result := sha256.Sum256([]byte(codeVerifier))
	enc := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_").WithPadding(base64.NoPadding)

	authParams := &AuthParams{
		ClientID:      h.ClientID,
		State:         traQutils.RandAlphabetAndNumberString(10),
		CodeChallenge: enc.EncodeToString(result[:]),
	}

	return c.JSON(http.StatusCreated, authParams)
}
