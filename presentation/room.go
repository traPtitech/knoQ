package presentation

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
)

//go:generate gotypeconverter -s RoomReq -d domain.WriteRoomParams -o converter.go .
type RoomReq struct {
	Place     string      `json:"place"`
	TimeStart time.Time   `json:"timeStart"`
	TimeEnd   time.Time   `json:"timeEnd"`
	Admins    []uuid.UUID `json:"admins"`
}

type RoomCSVReq struct {
	Subject   string `csv:"Subject"`
	StartDate string `csv:"Start date"`
	EndDate   string `csv:"End date"`
	StartTime string `csv:"Start time"`
	EndTime   string `csv:"End time"`
	Location  string `csv:"Location"`
}

type StartEndTime struct {
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

//go:generate gotypeconverter -s domain.Room -d RoomRes -o converter.go .
//go:generate gotypeconverter -s []*domain.Room -d []*RoomRes -o converter.go .
//go:generate gotypeconverter -s []domain.StartEndTime -d []StartEndTime -o converter.go .
type RoomRes struct {
	ID uuid.UUID `json:"roomId"`
	// Verifeid indicates if the room has been verified by privileged users.
	Verified      bool `json:"verified"`
	AllowTogether bool `json:"sharedRoom"`
	RoomReq
	FreeTimes   []StartEndTime `json:"freeTimes" cvt:"-"`
	SharedTimes []StartEndTime `json:"sharedTimes" cvt:"-"`
	CreatedBy   uuid.UUID      `json:"createdBy"`
	Model
}

func ChangeRoomCSVReqTodomainWriteRoomParams(src RoomCSVReq, userID uuid.UUID) (*domain.WriteRoomParams, error) {

	layout := "2006/01/02 15:04"
	jst, _ := time.LoadLocation("Asia/Tokyo")
	var params domain.WriteRoomParams
	var err error = nil

	params.Place = src.Location
	params.TimeStart, err = time.ParseInLocation(layout, src.StartDate+" "+src.StartTime, jst)
	if err != nil {
		return nil, err
	}

	params.TimeEnd, err = time.ParseInLocation(layout, src.EndDate+" "+src.EndTime, jst)
	if err != nil {
		return nil, err
	}

	params.Admins = []uuid.UUID{userID}

	return &params, err

}
