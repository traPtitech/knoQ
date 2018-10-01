package main

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
)

// GetHello テスト用API
func GetHello(c echo.Context) error {
	id := getRequestUser(c)                                      // リクエストしてきたユーザーのtraQID取得
	return c.String(http.StatusOK, fmt.Sprintf("hello %s!", id)) // レスポンスを返す
}
