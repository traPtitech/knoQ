package router

import (
	"net/http"
	repo "room/repository"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

// HandlePostGroup グループを作成
func HandlePostGroup(c echo.Context) error {
	g := new(repo.Group)

	if err := c.Bind(&g); err != nil {
		return badRequest(message(err.Error()))
	}

	g.CreatedBy = getRequestUser(c).TRAQID

	if err := g.Create(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return badRequest()
		}
		return internalServerError()
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
		if gorm.IsRecordNotFoundError(err) {
			return badRequest()
		}
		return internalServerError()
	}

	return c.NoContent(http.StatusOK)
}

// HandleUpdateGroup グループメンバー、グループ名を更新
func HandleUpdateGroup(c echo.Context) error {
	group := new(repo.Group)
	var err error
	if err := c.Bind(group); err != nil {
		return badRequest(message(err.Error()))
	}
	group.ID, err = getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}
	err = group.Update()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return badRequest()
		}
		return internalServerError()
	}

	return c.JSON(http.StatusOK, group)
}
