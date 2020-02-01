package router

import (
	"crypto/sha256"
	"fmt"
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

func HandlePostAuthParams(c echo.Context) error {
	authParams := new(AuthParams)
	codeVerifier := traQutils.RandAlphabetAndNumberString(43)

	// cache codeVerifier
	sess, _ := session.Get("session", c)
	sessionID := sess.ID
	verifierCache.Set(sessionID, codeVerifier, cache.DefaultExpiration)

	authParams = &AuthParams{
		ClientID:      "",
		State:         traQutils.RandAlphabetAndNumberString(10),
		CodeChallenge: fmt.Sprintf("%x", sha256.Sum256([]byte(codeVerifier))),
	}

	return c.JSON(http.StatusCreated, authParams)
}
