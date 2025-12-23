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

	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	group, err := h.Service.CreateGroup(ctx, reqID, groupParams)
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
	ctx := c.Request().Context()
	group, err := h.Service.GetGroup(ctx, groupID)
	if err != nil {
		return judgeErrorResponse(err)
	}

	res := presentation.ConvdomainGroupToGroupRes(*group)
	return c.JSON(http.StatusOK, res)
}

// HandleGetGroups グループを取得
func (h *Handlers) HandleGetGroups(c echo.Context) error {
	ctx := c.Request().Context()
	groups, err := h.Service.GetAllGroups(ctx)
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
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	if err := h.Service.DeleteGroup(ctx, reqID, groupID); err != nil {
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
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	res, err := h.Service.UpdateGroup(ctx, reqID, groupID, groupParams)
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
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	err = h.Service.AddMeToGroup(ctx, reqID, groupID)
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

	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	err = h.Service.DeleteMeGroup(ctx, reqID, groupID)
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleGetMeGroupIDs(c echo.Context) error {
	userID, _ := getRequestUserID(c)

	var groupIDs []uuid.UUID
	var err error

	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	switch presentation.GetUserRelationQuery(c.QueryParams()) {
	case presentation.RelationBelongs:
		groupIDs, err = h.Service.GetUserBelongingGroupIDs(ctx, reqID, userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
	case presentation.RelationAdmins:
		groupIDs, err = h.Service.GetUserAdminGroupIDs(ctx, userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
	case presentation.RelationBelongsOrAdmins:
		belongingGroupIDs, err := h.Service.GetUserBelongingGroupIDs(ctx, reqID, userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
		adminGroupIDs, err := h.Service.GetUserAdminGroupIDs(ctx, userID)
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
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	switch presentation.GetUserRelationQuery(c.QueryParams()) {
	case presentation.RelationBelongs:
		groupIDs, err = h.Service.GetUserBelongingGroupIDs(ctx, reqID, userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
	case presentation.RelationAdmins:
		groupIDs, err = h.Service.GetUserAdminGroupIDs(ctx, userID)
		if err != nil {
			return judgeErrorResponse(err)
		}
	}

	return c.JSON(http.StatusOK, groupIDs)
}
