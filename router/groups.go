package router

import (
	"fmt"
	"net/http"
	repo "room/repository"
	"strconv"

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

	if err := g.Delete(); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return notFound()
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

func HandleAddGroupTag(c echo.Context) error {
	tag := new(repo.Tag)
	group := new(repo.Group)
	if err := c.Bind(tag); err != nil {
		return badRequest()
	}
	var err error
	group.ID, err = getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}

	return handleAddTagRelation(c, group, group.ID, tag.Name)
}

func HandleDeleteGroupTag(c echo.Context) error {
	groupTag := new(repo.GroupTag)
	group := new(repo.Group)
	var err error
	group.ID, err = getRequestGroupID(c)
	if err != nil {
		internalServerError()
	}
	groupTag.TagID, err = strconv.ParseUint(c.Param("tagid"), 10, 64)
	if err != nil || groupTag.TagID == 0 {
		return notFound(message(fmt.Sprintf("TagID: %v does not exist.", c.Param("tagid"))))
	}

	return handleDeleteTagRelation(c, group, groupTag.TagID)
}
