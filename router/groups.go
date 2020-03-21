package router

import (
	"net/http"
	repo "room/repository"

	"github.com/jinzhu/copier"
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

	res := formatGroupRes(group, false)
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
			return internalServerError()
		}
		return c.JSON(http.StatusOK, formatGroupRes(group, true))
	}

	return c.JSON(http.StatusOK, formatGroupRes(group, false))
}

// HandleGetGroups グループを取得
func (h *Handlers) HandleGetGroups(c echo.Context) error {

	groups, err := h.Repo.GetAllGroups()
	if err != nil {
		return err
	}
	res := formatGroupsRes(groups, false)

	token, _ := getRequestUserToken(c)
	UserGroupRepo := h.InitExternalUserGroupRepo(token, repo.V3)
	traQgroups, err := UserGroupRepo.GetAllGroups()
	if err != nil {
		return err
	}
	res = append(res, formatGroupsRes(traQgroups, true)...)

	return c.JSON(http.StatusOK, res)
}

// HandleDeleteGroup グループを削除
func (h *Handlers) HandleDeleteGroup(c echo.Context) error {
	groupID, err := getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}

	if err := h.Repo.DeleteGroup(groupID); err != nil {
		return internalServerError()
	}

	return c.NoContent(http.StatusNoContent)
}

// HandleUpdateGroup 変更できるものはpostと同等
func (h *Handlers) HandleUpdateGroup(c echo.Context) error {
	g := new(GroupReq)
	if err := c.Bind(&g); err != nil {
		return badRequest(message(err.Error()))
	}
	groupParams := new(repo.WriteGroupParams)
	err := copier.Copy(&groupParams, g)
	if err != nil {
		return internalServerError()
	}

	groupID, err := getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}
	group, err := h.Repo.UpdateGroup(groupID, *groupParams)
	if err != nil {
		return internalServerError()
	}
	res := formatGroupRes(group, false)
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

func (h *Handlers) HandleAddMeGroup(c echo.Context) error {
	groupID, err := getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}

	userID, _ := getRequestUserID(c)
	if err := h.Repo.AddUserToGroup(groupID, userID); err != nil {
		return internalServerError()
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleDeleteMeGroup(c echo.Context) error {
	groupID, err := getRequestGroupID(c)
	if err != nil {
		return internalServerError()
	}

	userID, _ := getRequestUserID(c)
	if err := h.Repo.DeleteUserInGroup(groupID, userID); err != nil {
		return internalServerError()
	}

	return c.NoContent(http.StatusNoContent)
}
