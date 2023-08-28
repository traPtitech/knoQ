package router

import (
	"bytes"
	"net/http"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filter"
	"github.com/traPtitech/knoQ/presentation"
	"github.com/traPtitech/knoQ/utils/parsing"

	"github.com/gofrs/uuid"

	"github.com/labstack/echo/v4"
)

// HandlePostEvent 部屋の使用宣言を作成
func (h *Handlers) HandlePostEvent(c echo.Context) error {
	var req presentation.EventReqWrite
	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	params := presentation.ConvEventReqWriteTodomainWriteEventParams(req)

	event, err := h.Repo.CreateEvent(params, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusCreated, presentation.ConvdomainEventToEventDetailRes(*event))
}

// HandleUpdateEvent 任意の要素を変更
func (h *Handlers) HandleUpdateEvent(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err)
	}

	var req presentation.EventReqWrite
	if err := c.Bind(&req); err != nil {
		return badRequest(err, message(err.Error()))
	}
	params := presentation.ConvEventReqWriteTodomainWriteEventParams(req)

	event, err := h.Repo.UpdateEvent(eventID, params, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusCreated, presentation.ConvdomainEventToEventDetailRes(*event))
}

// HandleDeleteEvent 部屋の使用宣言を削除
func (h *Handlers) HandleDeleteEvent(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err)
	}

	if err = h.Repo.DeleteEvent(eventID, getConinfo(c)); err != nil {
		return internalServerError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// HandleGetEvent get one event
func (h *Handlers) HandleGetEvent(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err)
	}

	event, err := h.Repo.GetEvent(eventID, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvdomainEventToEventDetailRes(*event))
}

// HandleGetEvents 部屋の使用宣言情報を取得
func (h *Handlers) HandleGetEvents(c echo.Context) error {
	values := c.QueryParams()
	filterQuery := values.Get("q")
	expr, err := parsing.Parse(filterQuery)
	if err != nil {
		return badRequest(err, message("parse error"))
	}

	start, end, err := presentation.GetTiemRange(values)
	if err != nil {
		return badRequest(err, message("invalid time"))
	}
	events, err := h.Repo.GetEvents(
		filter.AddAnd(expr, filter.FilterTime(start, end)),
		getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.JSON(http.StatusOK, presentation.ConvSPdomainEventToSEventRes(events))
}

// HandleGetEventsByGroupID get events by groupID
// If groupID does not exist, this return []. Does not returns error.
func (h *Handlers) HandleGetEventsByGroupID(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}
	events, err := h.Repo.GetEvents(filter.FilterGroupIDs(groupID),
		getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvSPdomainEventToSEventRes(events))
}

func (h *Handlers) HandleAddEventTag(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err, message(err.Error()))
	}

	var req presentation.EventTagReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}

	err = h.Repo.AddEventTag(eventID, req.Name, false, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleDeleteEventTag(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err, message(err.Error()))
	}
	tagName := c.Param("tagName")

	err = h.Repo.DeleteEventTag(eventID, tagName, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleUpsertMeEventSchedule(c echo.Context) error {
	eventID, err := getPathEventID(c)
	if err != nil {
		return notFound(err, message(err.Error()))
	}

	var req presentation.EventScheduleStatusReq
	if err := c.Bind(&req); err != nil {
		return badRequest(err)
	}
	params := domain.ScheduleStatus(req.Schedule)

	err = h.Repo.UpsertMeEventSchedule(eventID, params, getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) HandleGetMeEvents(c echo.Context) error {
	userID, err := getRequestUserID(c)
	if err != nil {
		return notFound(err)
	}

	events, err := h.Repo.GetEvents(
		getUserRelationFilter(c.QueryParams(), userID),
		getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvSPdomainEventToSEventRes(events))
}

func (h *Handlers) HandleGetEventsByUserID(c echo.Context) error {
	userID, err := getPathUserID(c)
	if err != nil {
		return notFound(err)
	}

	events, err := h.Repo.GetEvents(
		getUserRelationFilter(c.QueryParams(), userID),
		getConinfo(c))
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvSPdomainEventToSEventRes(events))
}

// HandleGetEventsByRoomID get events by roomID
// If roomID does not exist, this return []. Does not returns error.
func (h *Handlers) HandleGetEventsByRoomID(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}

	events, err := h.Repo.GetEvents(
		filter.FilterRoomIDs(roomID),
		getConinfo(c),
	)
	if err != nil {
		return judgeErrorResponse(err)
	}
	return c.JSON(http.StatusOK, presentation.ConvSPdomainEventToSEventRes(events))
}

// HandleGetiCalByPrivateID sessionを持たないリクエストが想定されている
func (h *Handlers) HandleGetiCalByPrivateID(c echo.Context) error {
	// 認証
	str := c.Param("userIDsecret")
	userID, err := uuid.FromString(str[:36])
	if err != nil {
		return notFound(err)
	}
	info := &domain.ConInfo{ReqUserID: userID}
	icalSecret, err := h.Repo.GetMyiCalSecret(info)
	if err != nil {
		return judgeErrorResponse(err)
	}
	secret := str[36:]
	if icalSecret == "" || icalSecret != secret {
		return notFound(err)
	}

	filter := c.QueryParam("q")
	expr, err := parsing.Parse(filter)
	if err != nil {
		return badRequest(err)
	}
	events, err := h.Repo.GetEvents(expr, info)
	if err != nil {
		return judgeErrorResponse(err)
	}

	users, err := h.Repo.GetAllUsers(false, true, info)
	if err != nil {
		return judgeErrorResponse(err)
	}

	userMap := make(map[uuid.UUID]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	cal := presentation.ICalFormat(events, h.Origin, userMap)
	var buf bytes.Buffer
	_ = cal.SerializeTo(&buf)
	return c.Blob(http.StatusOK, "text/calendar", buf.Bytes())
}
