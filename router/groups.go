package router

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/router/presentation"

	"github.com/labstack/echo/v4"
)

// HandlePostGroup グループを作成
func (h *Handlers) HandlePostGroup(c echo.Context) error {
	var req presentation.GroupReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	groupParams := presentation.ConvGroupReqTodomainWriteGroupParams(req)

	group, err := h.Service.CreateGroup(groupParams, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	res := presentation.ConvdomainGroupToGroupRes(*group)
	return c.JSON(http.StatusCreated, res)
}

// HandleGetGroup グループを一件取得
func (h *Handlers) HandleGetGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	group, err := h.Service.GetGroup(groupID, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	res := presentation.ConvdomainGroupToGroupRes(*group)
	return c.JSON(http.StatusOK, res)
}

// HandleGetGroups グループを取得
func (h *Handlers) HandleGetGroups(c echo.Context) error {
	groups, err := h.Service.GetAllGroups(getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	res := presentation.ConvSPdomainGroupToSPGroupRes(groups)
	return c.JSON(http.StatusOK, res)
}

// HandleDeleteGroup グループを削除
func (h *Handlers) HandleDeleteGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	if err := h.Service.DeleteGroup(groupID, getConinfo(c)); err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// HandleUpdateGroup 変更できるものはpostと同等
func (h *Handlers) HandleUpdateGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	var req presentation.GroupReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	groupParams := presentation.ConvGroupReqTodomainWriteGroupParams(req)

	res, err := h.Service.UpdateGroup(groupID, groupParams, getConinfo(c))
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

	err = h.Service.AddMeToGroup(groupID, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleDeleteMeGroup(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	err = h.Service.DeleteMeGroup(groupID, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleGetMeGroupIDs(c echo.Context) error {
	userID, _ := getRequestUserID(c)

	var groupIDs []uuid.UUID
	var err error
	switch presentation.GetUserRelationQuery(c.QueryParams()) {
	case presentation.RelationBelongs:
		groupIDs, err = h.Service.GetUserBelongingGroupIDs(userID, getConinfo(c))
		if err != nil {
			return judgeErrorResponse(err)
		}
	case presentation.RelationAdmins:
		groupIDs, err = h.Service.GetUserAdminGroupIDs(userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
	case presentation.RelationBelongsOrAdmins:
		belongingGroupIDs, err := h.Service.GetUserBelongingGroupIDs(userID, getConinfo(c))
		if err != nil {
			return judgeErrorResponse(err)
		}
		adminGroupIDs, err := h.Service.GetUserAdminGroupIDs(userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
		allGroupIDs := belongingGroupIDs
		allGroupIDs = append(allGroupIDs, adminGroupIDs...)
		uniqueIDMap := make(map[uuid.UUID]struct{})

		for _, groupID := range allGroupIDs {
			if _, ok := uniqueIDMap[groupID]; ok {
				continue
			}

			uniqueIDMap[groupID] = struct{}{}
			groupIDs = append(groupIDs, groupID)
		}
	}

	return c.JSON(http.StatusOK, groupIDs)
}

func (h *Handlers) HandleGetGroupIDsByUserID(c echo.Context) error {
	userID, err := getPathUserID(c)
	if err != nil {
		return notFound(err, message(err.Error()))
	}
	var groupIDs []uuid.UUID
	switch presentation.GetUserRelationQuery(c.QueryParams()) {
	case presentation.RelationBelongs:
		groupIDs, err = h.Service.GetUserBelongingGroupIDs(userID, getConinfo(c))
		if err != nil {
			return judgeErrorResponse(err)
		}
	case presentation.RelationAdmins:
		groupIDs, err = h.Service.GetUserAdminGroupIDs(userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
	}

	return c.JSON(http.StatusOK, groupIDs)
}
