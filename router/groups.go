package router

import (
	"fmt"
	"net/http"
	"room/middleware"
	repo "room/repository"
	"strconv"

	"github.com/labstack/echo"
)

// HandlePostGroup グループを作成
func HandlePostGroup(c echo.Context) error {
	g := new(repo.Group)

	if err := c.Bind(&g); err != nil {
		return err
	}

	g.CreatedByRefer = middleware.GetRequestUser(c)
	if err := g.AddCreatedBy(); err != nil {
		return err
	}

	// メンバーがdbにいるか
	if err := g.FindMembers(); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	if err := repo.DB.Create(&g).Error; err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprint(err))
	}

	return c.JSON(http.StatusCreated, g)
}

// HandleGetGroups グループを取得
func HandleGetGroups(c echo.Context) error {
	groups := []repo.Group{}
	values := c.QueryParams()

	groups, err := repo.FindGroups(values)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, groups)
}

// HandleDeleteGroup グループを削除
func HandleDeleteGroup(c echo.Context) error {
	g := new(repo.Group)
	g.ID, _ = strconv.Atoi(c.Param("groupid"))

	fmt.Println(g.ID)
	if err := repo.DB.First(&g, g.ID).Related(&g.Members, "Members").Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	// relationを削除
	if err := repo.DB.Model(&g).Association("Members").Clear().Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	// 予約情報を削除
	if err := repo.DB.Where("group_id = ?", g.ID).Delete(&repo.Reservation{}).Error; err != nil {
		fmt.Println(err)
	}

	if err := repo.DB.Delete(&g).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

// HandleUpdateGroup グループメンバー、グループ名を更新
func HandleUpdateGroup(c echo.Context) error {
	g := new(repo.Group)

	if err := c.Bind(g); err != nil {
		return err
	}
	name := g.Name
	description := g.Description

	// メンバーがdbにいるか
	if err := g.FindMembers(); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	g.ID, _ = strconv.Atoi(c.Param("groupid"))
	if err := repo.DB.First(&g, g.ID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "アクセスしたgroupIDは存在しない")
	}
	// 作成者を取得
	if err := g.AddCreatedBy(); err != nil {
		return err
	}
	if middleware.GetRequestUser(c) != g.CreatedByRefer {
		return echo.NewHTTPError(http.StatusForbidden, "作成者ではない")
	}

	// メンバーを置き換え
	if err := repo.DB.Model(&g).Association("Members").Replace(g.Members).Error; err != nil {
		return err
	}

	// グループ名を変更
	if err := repo.DB.Model(&g).Update("name", name).Error; err != nil {
		return err
	}
	fmt.Println(g.Name)
	// グループ詳細変更
	if err := repo.DB.Model(&g).Update("description", description).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, g)
}
