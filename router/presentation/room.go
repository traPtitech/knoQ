package presentation

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/utils/tz"
)

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s RoomReq -d domain.WriteRoomParams -o converter.go .
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

//go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []domain.StartEndTime -d []StartEndTime -o converter.go .
type StartEndTime struct {
	TimeStart time.Time `json:"timeStart"`
	TimeEnd   time.Time `json:"timeEnd"`
}

type RoomRes struct {
	ID uuid.UUID `json:"roomId"`
	// Verifeid indicates if the room has been verified by privileged users.
	Verified bool `json:"verified"`
	RoomReq
	FreeTimes   []StartEndTime `json:"freeTimes" cvt:"-"`
	SharedTimes []StartEndTime `json:"sharedTimes" cvt:"-"`
	CreatedBy   uuid.UUID      `json:"createdBy"`
	Model
}

// TODO: FreeTimesとShareTimesを埋めるために手動で書いている
// //go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s []*domain.Room -d []*RoomRes -o converter.go .
func ConvSPdomainRoomToSPRoomRes(src []*domain.Room) (dst []*RoomRes) {
	dst = make([]*RoomRes, len(src))
	for i := range src {
		if src[i] != nil {
			dst[i] = new(RoomRes)
			(*dst[i]) = ConvdomainRoomToRoomRes((*src[i]))
		}
	}
	return
}

// //go:generate go run github.com/fuji8/gotypeconverter/cmd/gotypeconverter@latest -s domain.Room -d RoomRes -o converter.go .
func ConvdomainRoomToRoomRes(src domain.Room) (dst RoomRes) {
	dst.ID = src.ID
	dst.Verified = true // Room は全て進捗部屋
	dst.Place = src.Name
	dst.TimeStart = src.TimeStart
	dst.TimeEnd = src.TimeEnd
	dst.Admins = make([]uuid.UUID, len(src.Admins))
	for i := range src.Admins {
		dst.Admins[i] = convdomainUserTouuidUUID(src.Admins[i])
	}
	dst.CreatedBy = convdomainUserTouuidUUID(src.CreatedBy)
	dst.FreeTimes = ConvSdomainStartEndTimeToSStartEndTime(src.CalcAvailableTime(false))
	dst.SharedTimes = ConvSdomainStartEndTimeToSStartEndTime(src.CalcAvailableTime(true))
	dst.Model = Model(src.Model)
	return
}

func ChangeRoomCSVReqTodomainWriteRoomParams(src RoomCSVReq, userID uuid.UUID) (*domain.WriteRoomParams, error) {
	layout := "2006/01/02 15:04"
	var params domain.WriteRoomParams
	var err error

	params.Place = src.Location
	params.TimeStart, err = time.ParseInLocation(layout, src.StartDate+" "+src.StartTime, tz.JST)
	if err != nil {
		return nil, err
	}

	params.TimeEnd, err = time.ParseInLocation(layout, src.EndDate+" "+src.EndTime, tz.JST)
	if err != nil {
		return nil, err
	}

	params.Admins = []uuid.UUID{userID}

	return &params, err
}
