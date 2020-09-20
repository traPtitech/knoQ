package router

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"room/router/service"

	"github.com/labstack/echo/v4"
	traQv3 "github.com/traPtitech/traQ/router/v3"
	traQutils "github.com/traPtitech/traQ/utils"
)

// HandleGetUserMe ヘッダー情報からuser情報を取得
// 認証状態を確認
func (h *Handlers) HandleGetUserMe(c echo.Context) error {
	token, _ := getRequestUserToken(c)
	userID, _ := getRequestUserID(c)

	user, err := h.Dao.GetUser(token, userID)
	if err != nil {
		if err.Error() == http.StatusText(http.StatusUnauthorized) {
			h.Repo.ReplaceToken(userID, "")
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, service.FormatUserRes(user))
}

// HandleGetUsers ユーザーすべてを取得
func (h *Handlers) HandleGetUsers(c echo.Context) error {
	token, _ := getRequestUserToken(c)
	userID, _ := getRequestUserID(c)

	users, err := h.Dao.GetAllUsers(token)
	if err != nil {
		if err.Error() == http.StatusText(http.StatusUnauthorized) {
			h.Repo.ReplaceToken(userID, "")
			return forbidden(err, message("token is invalid."), needAuthorization(true))
		}
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, service.FormatUsersRes(users))
}
func (h *Handlers) HandleGetiCal(c echo.Context) error {
	userID, _ := getRequestUserID(c)
	secret, err := h.Dao.Repo.GetiCalSecret(userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, struct {
		Secret string `json:"secret"`
	}{
		Secret: secret,
	})
}

func (h *Handlers) HandleUpdateiCal(c echo.Context) error {
	userID, _ := getRequestUserID(c)
	secret := traQutils.RandAlphabetAndNumberString(16)
	if err := h.Dao.Repo.ReplaceiCalSecret(userID, secret); err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, struct {
		Secret string `json:"secret"`
	}{
		Secret: secret,
	})
}

// HandleSyncUser traQのユーザーとの同期をする
// 停止されているユーザーの`token`を削除して、
// 活動中のユーザーを追加する(userIDをDBに保存)
func (h *Handlers) HandleSyncUser(c echo.Context) error {
	token, _ := getRequestUserToken(c)

	// TODO fix with repository v4
	allUsers, err := getTraQAllUsers(token)
	if err != nil {
		return internalServerError(err)
	}
	for _, user := range allUsers {
		if user.State == 0 {
			h.Repo.ReplaceToken(user.ID, "")
		} else if user.State == 1 {
			h.Repo.SaveUser(user.ID, false, true)
		}
	}
	return c.NoContent(http.StatusCreated)
}

func getTraQAllUsers(token string) ([]traQv3.User, error) {
	u, err := url.Parse("https://q.trap.jp/api/v3")
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "users")
	query := u.Query()
	query.Set("include-suspended", "1")
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	users := make([]traQv3.User, 0)
	err = json.Unmarshal(data, &users)
	return users, err
}
