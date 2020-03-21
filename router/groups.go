package router

import (
	"fmt"
	"net/http"
	repo "room/repository"

	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

// HandlePostGroup グループを作成
func (h *Handlers) HandlePostGroup(c echo.Context) error {
	g := new(GroupReq)

	if err := c.Bind(&g); err != nil {
		return badRequest(message(err.Error()))
	}
	groupParams := new(repo.WriteGroupParams)
	err := copier.Copy(&groupParams, g)
	if err != nil {
		return internalServerError()
	}

	groupParams.CreatedBy, _ = getRequestUserID(c)

	group, err := h.Repo.CreateGroup(*groupParams)
	if err != nil {
		return internalServerError()
	}

	res, err := formatGroupRes(group)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusCreated, res)
}

// HandleGetGroup グループを一件取得
// TODO fix
func (h *Handlers) HandleGetGroup(c echo.Context) error {
	groupID, err := getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}

	group, _ := h.Repo.GetGroup(groupID)
	if group == nil {
		token, _ := getRequestUserToken(c)
		UserGroupRepo := h.InitExternalUserGroupRepo(token, repo.V3)
		group, err = UserGroupRepo.GetGroup(groupID)
		if err != nil {
			fmt.Println(err)
			return internalServerError()
		}
	}

	res, err := formatGroupRes(group)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, res)
}

// HandleGetGroups グループを取得
func HandleGetGroups(c echo.Context) error {
	values := c.QueryParams()

	groups, err := repo.FindGroups(values)
	if err != nil {
		return err
	}
	res, err := formatGroupsRes(groups)
	if err != nil {
		return internalServerError()
	}

	return c.JSON(http.StatusOK, res)
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

	res, err := formatGroupRes(group)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, res)
}

/*
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
		return internalServerError()
	}
	groupTag.TagID, err = uuid.FromString(c.Param("tagid"), 10, 64)
	if err != nil || groupTag.TagID == 0 {
		return notFound(message(fmt.Sprintf("TagID: %v does not exist.", c.Param("tagid"))))
	}

	return handleDeleteTagRelation(c, group, groupTag.TagID)
}
*/

func HandleAddMeGroup(c echo.Context) error {
	group := new(repo.Group)
	var err error
	group.ID, err = getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}
	if err := group.Read(); err != nil {
		return internalServerError()
	}
	if !group.JoinFreely {
		return forbidden(message("This group is not JoinFreely."), specification("This api can delete me at JoinFreely-group."))
	}

	userID, _ := getRequestUserID(c)
	if err := group.AddMember(userID); err != nil {
		return judgeErrorResponse(err)
	}
	if err := group.Read(); err != nil {
		return internalServerError()
	}

	res, err := formatGroupRes(group)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, res)
}

func HandleDeleteMeGroup(c echo.Context) error {
	group := new(repo.Group)
	var err error
	group.ID, err = getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}
	if err := group.Read(); err != nil {
		return internalServerError()
	}
	if !group.JoinFreely {
		return forbidden(message("This group is not JoinFreely."), specification("This api can delete me at JoinFreely-group."))
	}

	userID, _ := getRequestUserID(c)
	if err := group.DeleteMember(userID); err != nil {
		return judgeErrorResponse(err)
	}

	if err := group.Read(); err != nil {
		return internalServerError()
	}
	res, err := formatGroupRes(group)
	if err != nil {
		return internalServerError()
	}
	return c.JSON(http.StatusOK, res)
}
