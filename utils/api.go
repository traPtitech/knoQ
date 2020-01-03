package utils

import (
	"errors"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://q.trap.jp/api/1.0"

func GetUserMe(token string) ([]byte, error) {
	if token == "" {
		return nil, errors.New(http.StatusText(http.StatusUnauthorized))
	}
	req, err := http.NewRequest(http.MethodGet, baseURL+"users/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}
	return ioutil.ReadAll(res.Body)
}
