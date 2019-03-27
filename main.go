package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	MARIADB_HOSTNAME = os.Getenv("MARIADB_HOSTNAME")
	MARIADB_DATABASE = os.Getenv("MARIADB_DATABASE")
	MARIADB_USERNAME = os.Getenv("MARIADB_USERNAME")
	MARIADB_PASSWORD = os.Getenv("MARIADB_PASSWORD")

	db *gorm.DB
)

func main() {
	var err error

	// データベース接続
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", MARIADB_USERNAME, MARIADB_PASSWORD, MARIADB_HOSTNAME, MARIADB_DATABASE))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	// echo初期化
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Secure())
	e.Use(middleware.Static("./web/dist"))

	// API定義 (/api)
	api := e.Group("/api", traQUserMiddleware)
	api.GET("/hello", GetHello) // テスト用
	api.GET("/rooms", GetRooms)
	api.DELETE("/rooms/:roomid", DeleteRoom)

	// 管理者専用API定義 (/api/admin)
	adminApi := api.Group("/admin", adminUserMiddleware)
	adminApi.GET("/hello", GetHello) // テスト用
	adminApi.POST("/rooms", SaveRoom)

	// サーバースタート
	go func() {
		if err := e.Start(":3000"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

// initDB データベースのスキーマを更新
func initDB() error {
	// テーブルが無ければ作成
	if err := db.AutoMigrate(tables...).Error; err != nil {
		return err
	}
	return nil
}
