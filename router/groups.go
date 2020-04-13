package router

import (
	"net/http"
	repo "room/repository"
	"room/router/service"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
)

// HandlePostGroup グループを作成
func (h *Handlers) HandlePostGroup(c echo.Context) error {
	var req service.GroupReq

	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	groupParams := new(repo.WriteGroupParams)
	err := copier.Copy(&groupParams, req)
	if err != nil {
		return internalServerError(err)
	}
	groupParams.CreatedBy, _ = getRequestUserID(c)

	token, _ := getRequestUserToken(c)

	res, err := h.Dao.CreateGroup(token, *groupParams)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusCreated, res)
}

// HandleGetGroup グループを一件取得
// TODO fix
func (h *Handlers) HandleGetGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	token, _ := getRequestUserToken(c)
	groupRes, err := h.Dao.GetGroup(token, groupID)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, groupRes)
}

// HandleGetGroups グループを取得
func (h *Handlers) HandleGetGroups(c echo.Context) error {

	groups, err := h.Repo.GetAllGroups()
	if err != nil {
		return judgeErrorResponse(err)
	}
	res := service.FormatGroupsRes(groups, false)

	token, _ := getRequestUserToken(c)
	UserGroupRepo := h.InitExternalUserGroupRepo(token, repo.V3)
	traQgroups, err := UserGroupRepo.GetAllGroups()
	if err != nil {
		return judgeErrorResponse(err)
	}
	res = append(res, service.FormatGroupsRes(traQgroups, true)...)

	return c.JSON(http.StatusOK, res)
}

// HandleDeleteGroup グループを削除
func (h *Handlers) HandleDeleteGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	if err := h.Repo.DeleteGroup(groupID); err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// HandleUpdateGroup 変更できるものはpostと同等
func (h *Handlers) HandleUpdateGroup(c echo.Context) error {
	var req service.GroupReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	groupParams := new(repo.WriteGroupParams)
	err := copier.Copy(&groupParams, req)
	if err != nil {
		return internalServerError(err)
	}

	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}
	token, _ := getRequestUserToken(c)
	groupParams.CreatedBy, _ = getRequestUserID(c)
	res, err := h.Dao.UpdateGroup(token, groupID, *groupParams)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handlers) HandleAddMeGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	userID, _ := getRequestUserID(c)
	token, _ := getRequestUserToken(c)
	if err := h.Dao.AddUserToGroup(token, groupID, userID); err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleDeleteMeGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	userID, _ := getRequestUserID(c)
	if err := h.Repo.DeleteUserInGroup(groupID, userID); err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleGetMeGroups(c echo.Context) error {
	userID, _ := getRequestUserID(c)

	token, _ := getRequestUserToken(c)
	groupIDs, err := h.Dao.GetUserBelongingGroupIDs(token, userID)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, groupIDs)
}
