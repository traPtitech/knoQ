package router

import (
	"bytes"
	"net/http"

	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/domain/filters"
	"github.com/traPtitech/knoQ/router/presentation"
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
	ctx := c.Request().Context()
	reqID := c.Get(userIDKey).(uuid.UUID)
	event, err := h.Service.CreateEvent(ctx, reqID, params)
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

	reqID := c.Get(userIDKey).(uuid.UUID)
	event, err := h.Service.UpdateEvent(c.Request().Context(), reqID, eventID, params)
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

	reqID := c.Get(userIDKey).(uuid.UUID)
	if err = h.Service.DeleteEvent(c.Request().Context(), reqID, eventID); err != nil {
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

	event, err := h.Service.GetEvent(c.Request().Context(), eventID)
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
	durationExpr, err := getDurationFilter(values)
	if err != nil {
		return badRequest(err, message("filter duration error"))
	}

	reqID := c.Get(userIDKey).(uuid.UUID)
	combinedExpr := filters.AddAnd(expr, durationExpr)

	events, err := h.Service.GetEvents(c.Request().Context(), reqID, combinedExpr)
	if err != nil {
		return judgeErrorResponse(err)
	}

	eventsRes := presentation.ConvDomainEventsToEventsResElems(events)
	return c.JSON(http.StatusOK, eventsRes)
}

// HandleGetEventsByGroupID get events by groupID
// If groupID does not exist, this return []. Does not returns error.
func (h *Handlers) HandleGetEventsByGroupID(c echo.Context) error {
	groupID, err := getPathGroupID(c)
	if err != nil {
		return notFound(err)
	}

	values := c.QueryParams()

	groupExpr := filters.FilterGroupIDs(groupID)

	durationExpr, err := getDurationFilter(values)
	if err != nil {
		return badRequest(err, message("filter duration error"))
	}

	reqID := c.Get(userIDKey).(uuid.UUID)
	combinedExpr := filters.AddAnd(groupExpr, durationExpr)

	events, err := h.Service.GetEvents(c.Request().Context(), reqID, combinedExpr)
	if err != nil {
		return judgeErrorResponse(err)
	}
	eventsRes := presentation.ConvDomainEventsToEventsResElems(events)
	return c.JSON(http.StatusOK, eventsRes)
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

	reqID := c.Get(userIDKey).(uuid.UUID)
	err = h.Service.AddEventTag(c.Request().Context(), reqID, eventID, req.Name, false)
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

	reqID := c.Get(userIDKey).(uuid.UUID)
	tagName := c.Param("tagName")

	err = h.Service.DeleteEventTag(c.Request().Context(), reqID, eventID, tagName)
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
	reqID := c.Get(userIDKey).(uuid.UUID)

	err = h.Service.UpsertMeEventSchedule(c.Request().Context(), reqID, eventID, params)
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

	values := c.QueryParams()

	relationExpr := getUserRelationFilter(values, userID)

	durationExpr, err := getDurationFilter(values)
	if err != nil {
		return badRequest(err, message("filter duration error"))
	}

	reqID := c.Get(userIDKey).(uuid.UUID)
	combinedExpr := filters.AddAnd(relationExpr, durationExpr)

	events, err := h.Service.GetEvents(c.Request().Context(), reqID, combinedExpr)
	if err != nil {
		return judgeErrorResponse(err)
	}
	eventsRes := presentation.ConvDomainEventsToEventsResElems(events)
	return c.JSON(http.StatusOK, eventsRes)
}

func (h *Handlers) HandleGetEventsByUserID(c echo.Context) error {
	userID, err := getPathUserID(c)
	if err != nil {
		return notFound(err)
	}

	values := c.QueryParams()

	relationExpr := getUserRelationFilter(values, userID)

	durationExpr, err := getDurationFilter(values)
	if err != nil {
		return badRequest(err, message("filter duration error"))
	}

	reqID := c.Get(userIDKey).(uuid.UUID)

	combinedExpr := filters.AddAnd(relationExpr, durationExpr)

	events, err := h.Service.GetEvents(c.Request().Context(), reqID, combinedExpr)
	if err != nil {
		return judgeErrorResponse(err)
	}
	eventsRes := presentation.ConvDomainEventsToEventsResElems(events)
	return c.JSON(http.StatusOK, eventsRes)
}

// HandleGetEventsByRoomID get events by roomID
// If roomID does not exist, this return []. Does not returns error.
func (h *Handlers) HandleGetEventsByRoomID(c echo.Context) error {
	roomID, err := getPathRoomID(c)
	if err != nil {
		return notFound(err)
	}

	values := c.QueryParams()

	roomExpr := filters.FilterRoomIDs(roomID)

	durationExpr, err := getDurationFilter(values)
	if err != nil {
		return badRequest(err, message("filter duration error"))
	}

	reqID := c.Get(userIDKey).(uuid.UUID)
	combinedExpr := filters.AddAnd(roomExpr, durationExpr)

	events, err := h.Service.GetEvents(
		c.Request().Context(),
		reqID,
		combinedExpr,
	)
	if err != nil {
		return judgeErrorResponse(err)
	}
	eventsRes := presentation.ConvDomainEventsToEventsResElems(events)
	return c.JSON(http.StatusOK, eventsRes)
}

// HandleGetiCalByPrivateID sessionを持たないリクエストが想定されている
func (h *Handlers) HandleGetiCalByPrivateID(c echo.Context) error {
	// 認証
	str := c.Param("userIDsecret")
	userID, err := uuid.FromString(str[:36])
	if err != nil {
		return notFound(err)
	}

	ctx := c.Request().Context()
	icalSecret, err := h.Service.GetMyiCalSecret(ctx, userID)
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

	events, err := h.Service.GetEventsWithGroup(c.Request().Context(), userID, expr)
	if err != nil {
		return judgeErrorResponse(err)
	}

	users, err := h.Service.GetAllUsers(ctx, false, true)
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
