package service

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

func (s *service) CreateUnVerifiedRoom(ctx context.Context, reqID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	if !params.TimeConsistency() {
		return nil, ErrTimeConsistency
	}
	p := domain.CreateRoomArgs{
		WriteRoomParams: params,
		Verified:        false,
		CreatedBy:       reqID,
	}
	var roomResp *domain.Room
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		var err error
		roomResp, err = s.GormRepo.CreateRoom(ctx, p)
		return err
	})
	return roomResp, defaultErrorHandling(err)
}

func (s *service) CreateVerifiedRoom(ctx context.Context, reqID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	if !s.IsPrivilege(ctx, reqID) {
		return nil, domain.ErrForbidden
	}
	if !params.TimeConsistency() {
		return nil, ErrTimeConsistency
	}
	p := domain.CreateRoomArgs{
		WriteRoomParams: params,
		Verified:        true,
		CreatedBy:       reqID,
	}

	var roomResp *domain.Room
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		var err error
		roomResp, err = s.GormRepo.CreateRoom(ctx, p)
		return err
	})
	return roomResp, defaultErrorHandling(err)
}

func (s *service) UpdateRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID, params domain.WriteRoomParams) (*domain.Room, error) {
	if roomID == uuid.Nil {
		return nil, ErrRoomUndefined
	}
	if !s.IsRoomAdmins(ctx, reqID, roomID) {
		return nil, domain.ErrForbidden
	}

	p := domain.CreateRoomArgs{
		WriteRoomParams: params,
		Verified:        true,
		CreatedBy:       reqID,
	}

	if !params.TimeConsistency() {
		return nil, ErrTimeConsistency
	}

	var roomResp *domain.Room
	var err error
	room, err3 := s.GormRepo.GetRoom(ctx, roomID, uuid.Nil)
	if err3 != nil {
		return nil, defaultErrorHandling(err3)
	}
	err2 := s.GormRepo.DeleteRoom(ctx, roomID)
	if err2 != nil {
		return nil, defaultErrorHandling(err2)
	}
	roomResp, err = s.GormRepo.CreateRoom(ctx, p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	// oldRoom でのイベントの場所をすべて roomResp.ID に置換する
	events := room.Events
	for _, e := range events {
		ead := e.Admins
		admins := make([]uuid.UUID, len(ead))
		for i, a := range ead {
			admins[i] = a.ID
		}
		eta := e.Tags
		tags := make([]domain.EventTagParams, len(eta))
		for i, t := range eta {
			tags[i] = domain.EventTagParams{
				Name:   t.Tag.Name,
				Locked: t.Locked,
			}
		}
		_, err4 := s.GormRepo.UpdateEvent(ctx, e.ID, domain.UpsertEventArgs{
			WriteEventParams: domain.WriteEventParams{
				Name:          e.Name,
				Description:   e.Description,
				GroupID:       e.Group.ID,
				RoomID:        roomResp.ID,
				Place:         e.Room.Place,
				TimeStart:     e.TimeStart,
				TimeEnd:       e.TimeEnd,
				Admins:        admins,
				Tags:          tags,
				AllowTogether: e.AllowTogether,
				Open:          e.Open,
			},
			CreatedBy: e.CreatedBy.ID,
		})
		if err4 != nil {
			return nil, defaultErrorHandling(err4)
		}
	}
	return roomResp, defaultErrorHandling(err)
}

func (s *service) VerifyRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsPrivilege(ctx, reqID) {
		return domain.ErrForbidden
	}

	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		err := s.GormRepo.UpdateRoomVerified(ctx, roomID, true)
		return err
	})

	return defaultErrorHandling(err)
}

func (s *service) UnVerifyRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsPrivilege(ctx, reqID) {
		return domain.ErrForbidden
	}
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		err := s.GormRepo.UpdateRoomVerified(ctx, roomID, false)
		return err
	})

	return defaultErrorHandling(err)
}

func (s *service) DeleteRoom(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) error {
	if !s.IsRoomAdmins(ctx, reqID, roomID) {
		return domain.ErrForbidden
	}
	err := s.TxManager.Do(ctx, func(ctx context.Context) error {
		err := s.GormRepo.DeleteRoom(ctx, roomID)
		return err
	})

	return defaultErrorHandling(err)
}

func (s *service) GetRoom(ctx context.Context, roomID uuid.UUID, excludeEventID uuid.UUID) (*domain.Room, error) {
	rs, err := s.GormRepo.GetRoom(ctx, roomID, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (s *service) GetAllRooms(ctx context.Context, start time.Time, end time.Time, excludeEventID uuid.UUID) ([]*domain.Room, error) {
	rs, err := s.GormRepo.GetAllRooms(ctx, start, end, excludeEventID)
	return rs, defaultErrorHandling(err)
}

func (s *service) IsRoomAdmins(ctx context.Context, reqID uuid.UUID, roomID uuid.UUID) bool {
	room, err := s.GetRoom(ctx, roomID, uuid.Nil)
	if err != nil {
		return false
	}
	for _, admin := range room.Admins {
		if reqID == admin.ID {
			return true
		}
	}
	return false
}
