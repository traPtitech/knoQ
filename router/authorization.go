package router

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	traQutils "github.com/traPtitech/traQ/utils"
)

type AuthParams struct {
	ClientID      string `json:"clientId"`
	State         string `json:"state"`
	CodeChallenge string `json:"codeChallenge"`
}

func HandlePostAuthParams(c echo.Context) error {
	authParams := new(AuthParams)
	codeVerifier := traQutils.RandAlphabetAndNumberString(43)
	// Todo cache codeVerifier

	authParams = &AuthParams{
		ClientID:      "",
		State:         traQutils.RandAlphabetAndNumberString(10),
		CodeChallenge: fmt.Sprintf("%x", sha256.Sum256([]byte(codeVerifier))),
	}

	return c.JSON(http.StatusCreated, authParams)
}
