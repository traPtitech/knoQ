package repository

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/api/calendar/v3"
)

type WriteRoomParams struct {
	Place     string
	Public    bool
	TimeStart time.Time
	TimeEnd   time.Time
	CreatedBy uuid.UUID
}

// RoomRepository is implemted GormRepositoty and API repository
type RoomRepository interface {
	CreateRoom(roomParams WriteRoomParams) (*Room, error)
	UpdateRoom(roomID uuid.UUID, roomParams WriteRoomParams) (*Room, error)
	DeleteRoom(roomID uuid.UUID, deletePublic bool) error
	GetRoom(roomID uuid.UUID) (*Room, error)
	GetAllRooms(start *time.Time, end *time.Time) ([]*Room, error)
}

// BeforeCreate is gorm hook
func (r *Room) BeforeCreate() (err error) {
	r.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	return nil
}

func (repo *GormRepository) CreateRoom(roomParams WriteRoomParams) (*Room, error) {
	room := new(Room)
	err := copier.Copy(&room, roomParams)
	if err != nil {
		return nil, err
	}
	if !room.isTimeContext() {
		return nil, ErrInvalidArg
	}
	result := repo.DB.Debug().Set("gorm:insert_option", "ON DUPLICATE KEY UPDATE updated_at=updated_at").Create(&room)
	if result.RowsAffected == 0 {
		// duplicate
		room = new(Room)
		copier.Copy(&room, roomParams)
		room.TimeStart = room.TimeStart.UTC().Truncate(time.Second)
		room.TimeEnd = room.TimeEnd.UTC().Truncate(time.Second)
		if err := repo.DB.Debug().Where(&room).Where("public = ?", room.Public).Take(&room).Error; err != nil {
			return nil, err
		}
	} else if result.Error != nil {
		return nil, result.Error
	}

	return room, nil
}

func (repo *GormRepository) UpdateRoom(roomID uuid.UUID, roomParams WriteRoomParams) (*Room, error) {
	if roomID == uuid.Nil {
		return nil, ErrNilID
	}
	room := new(Room)
	err := copier.Copy(&room, roomParams)
	room.ID = roomID
	if err != nil {
		return nil, err
	}
	if !room.isTimeContext() {
		return nil, ErrInvalidArg
	}
	err = repo.DB.Save(&room).Error
	return room, err
}

func (repo *GormRepository) DeleteRoom(roomID uuid.UUID, deletePublic bool) error {
	if roomID == uuid.Nil {
		return ErrNilID
	}
	cmd := repo.DB
	if !deletePublic {
		cmd = cmd.Where("public = ?", false)
	}
	result := cmd.Delete(&Room{ID: roomID})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil

}
func (repo *GormRepository) GetRoom(roomID uuid.UUID) (*Room, error) {
	if roomID == uuid.Nil {
		return nil, ErrNilID
	}
	room := new(Room)
	room.ID = roomID
	err := repo.DB.Preload("Events").Take(&room).Error
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (repo *GormRepository) GetAllRooms(start *time.Time, end *time.Time) ([]*Room, error) {
	rooms := make([]*Room, 0)
	cmd := repo.DB.Preload("Events")
	if start != nil && !start.IsZero() {
		cmd = cmd.Where("time_start >= ?", start.UTC())
	}
	if end != nil && !end.IsZero() {
		cmd = cmd.Where("time_end <= ?", end.UTC())
	}
	err := cmd.Debug().Order("time_start").Find(&rooms).Error
	return rooms, err
}

func (repo *GoogleAPIRepository) CreateRoom(roomParams WriteRoomParams) (*Room, error) {
	return nil, ErrForbidden
}

func (repo *GoogleAPIRepository) UpdateRoom(roomID uuid.UUID, roomParams WriteRoomParams) (*Room, error) {
	return nil, ErrForbidden
}

func (repo *GoogleAPIRepository) DeleteRoom(roomID uuid.UUID, deletePublic bool) error {
	return nil
}

func (repo *GoogleAPIRepository) GetRoom(roomID uuid.UUID) (*Room, error) {
	return nil, ErrForbidden
}

func (repo *GoogleAPIRepository) GetAllRooms(start *time.Time, end *time.Time) ([]*Room, error) {
	srv, err := calendar.New(repo.Client)
	if err != nil {
		return nil, err
	}
	cmd := srv.Events.List(repo.CalendarID).ShowDeleted(false).SingleEvents(true)
	if start != nil {
		cmd = cmd.TimeMin(start.Format(time.RFC3339))
	}
	if end != nil {
		cmd = cmd.TimeMax(end.Format(time.RFC3339))
	}

	events, err := cmd.OrderBy("startTime").Do()
	if err != nil {
		return nil, err
	}
	return formatCalendar(events)
}

func formatCalendar(events *calendar.Events) ([]*Room, error) {
	if len(events.Items) == 0 {
		return nil, nil
	}
	rooms := make([]*Room, 0)
	for _, item := range events.Items {
		var err error
		room := &Room{
			Place:  item.Location,
			Public: true,
		}
		room.TimeStart, err = time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			return nil, err
		}

		room.TimeEnd, err = time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			return nil, err
		}
		room.TimeStart = room.TimeStart.UTC()
		room.TimeEnd = room.TimeEnd.UTC()

		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (room *Room) InTime(targetStartTime, targetEndTime time.Time, allowTogether bool) bool {
	for _, v := range room.CalcAvailableTime(allowTogether) {
		roomStart := v.TimeStart
		roomEnd := v.TimeEnd
		if (roomStart.Equal(targetStartTime) || roomStart.Before(targetStartTime)) && (roomEnd.Equal(targetEndTime) || roomEnd.After(targetEndTime)) {
			return true
		}
	}
	return false
}

// CalcAvailableTime calclate available time
// allowTogether = true 併用化の時間帯
// allowTogether = false 誰も取っていない時間帯
func (r *Room) CalcAvailableTime(allowTogether bool) []StartEndTime {
	availabletime := []StartEndTime{
		{
			TimeStart: r.TimeStart,
			TimeEnd:   r.TimeEnd,
		},
	}
	for _, e := range r.Events {
		if allowTogether && e.AllowTogether {
			continue
		}
		TimeRangesSub(availabletime, StartEndTime{e.TimeStart, e.TimeEnd})
	}
	return availabletime
}

func TimeRangesSub(as []StartEndTime, b StartEndTime) (cs []StartEndTime) {
	for _, a := range as {
		cs = append(cs, TimeRangeSub(a, b)...)
	}
	return
}

func TimeRangeSub(a StartEndTime, b StartEndTime) []StartEndTime {
	/*
		a: s####e-------
		b: -------s####e
		-> s####e
	*/
	if a.TimeStart.Unix() >= b.TimeEnd.Unix() || a.TimeEnd.Unix() <= b.TimeEnd.Unix() {
		return []StartEndTime{a}
	}

	/*
		a: ---s#####e---
		b: s##########e-
		-> -------------
	*/
	if b.TimeStart.Unix() <= a.TimeStart.Unix() && b.TimeEnd.Unix() >= a.TimeEnd.Unix() {
		return nil
	}

	/*
		a: s###########e
		b: ----s####e---
		-> s###e----s##e
	*/
	if a.TimeStart.Unix() < b.TimeStart.Unix() && b.TimeEnd.Unix() < a.TimeEnd.Unix() {
		return []StartEndTime{
			{a.TimeStart, b.TimeStart},
			{b.TimeEnd, a.TimeEnd},
		}
	}

	/*
		a: s#####e------
		b: ----s######e-
		-> s###e--------
	*/
	if a.TimeStart.Unix() < b.TimeStart.Unix() && a.TimeEnd.Unix() < b.TimeEnd.Unix() {
		return []StartEndTime{
			{a.TimeStart, b.TimeStart},
		}
	}

	/*
		a: -----s######e
		b: --s#####e----
		-> --------s###e
	*/
	if b.TimeStart.Unix() < a.TimeStart.Unix() && b.TimeEnd.Unix() < a.TimeEnd.Unix() {
		return []StartEndTime{
			{b.TimeEnd, a.TimeEnd},
		}
	}
	return nil
}

// isTimeContext 開始時間が終了時間より前か見る
func (r *Room) isTimeContext() bool {
	return r.TimeStart.Before(r.TimeEnd)
}
