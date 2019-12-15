package router

import (
	"fmt"
	"net/http"
	repo "room/repository"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

// HandlePostGroup グループを作成
func HandlePostGroup(c echo.Context) error {
	g := new(repo.Group)

	if err := c.Bind(&g); err != nil {
		return err
	}

	g.CreatedBy = getRequestUser(c).TRAQID

	// メンバーがdbにいるか
	if err := g.FindMembers(); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	if err := repo.DB.Create(&g).Error; err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprint(err))
	}

	return c.JSON(http.StatusCreated, g)
}

// HandleGetGroup グループを一件取得
func HandleGetGroup(c echo.Context) error {
	group := new(repo.Group)
	var err error
	group.ID, err = getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}
	if err := group.Read(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return notFound()
		}
		return internalServerError()
	}
	return c.JSON(http.StatusOK, group)
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
	var err error
	g.ID, err = getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}

	if err := repo.DB.Delete(&g).Error; err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

// HandleUpdateGroup グループメンバー、グループ名を更新
func HandleUpdateGroup(c echo.Context) error {
	g := new(repo.Group)
	var err error

	if err := c.Bind(g); err != nil {
		return err
	}
	name := g.Name
	description := g.Description

	// メンバーがdbにいるか
	if err := g.FindMembers(); err != nil {
		return c.String(http.StatusBadRequest, "正しくないメンバーが含まれている")
	}

	g.ID, err = getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}

	// メンバーを置き換え
	if err := repo.DB.Model(&g).Association("Members").Replace(g.Members).Error; err != nil {
		return err
	}

	// グループ名を変更
	if err := repo.DB.Model(&g).Update("name", name).Error; err != nil {
		return err
	}
	// グループ詳細変更
	if err := repo.DB.Model(&g).Update("description", description).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, g)
}
