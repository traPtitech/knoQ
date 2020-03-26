package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type WriteRoomParams struct {
	Place     string
	TimeStart time.Time
	TimeEnd   time.Time
}

// RoomRepository is implemted GormRepositoty and API repository
type RoomRepository interface {
	CreateRoom(roomParams WriteRoomParams) (*Room, error)
	UpdateRoom(roomID uuid.UUID, roomParams WriteRoomParams) (*Room, error)
	DeleteRoom(roomID uuid.UUID) error
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
	err = repo.DB.Create(&room).Error
	return room, err
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

func (repo *GormRepository) DeleteRoom(roomID uuid.UUID) error {
	if roomID == uuid.Nil {
		return ErrNilID
	}
	result := repo.DB.Delete(&Room{ID: roomID})
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

	room.calcAvailableTime()
	return room, nil
}

func (repo *GormRepository) GetAllRooms(start *time.Time, end *time.Time) ([]*Room, error) {
	rooms := make([]*Room, 0)
	cmd := repo.DB
	if start != nil {
		cmd = cmd.Where("rooms.end_time >= ?", start.String())
	}
	if end != nil {
		cmd = cmd.Where("rooms.start_time <= ?", end.String())
	}
	err := cmd.Find(&rooms).Error
	return rooms, err
}

//func FindRooms(values url.Values) ([]Room, error) {
//rooms := []Room{}
//cmd := DB
//if values.Get("dateBegin") != "" {
//cmd = cmd.Where("rooms.date >= ?", values.Get("dateBegin"))
//}
//if values.Get("dateEnd") != "" {
//cmd = cmd.Where("rooms.date <= ?", values.Get("dateEnd"))
//}

//rows, err := cmd.Order("date asc").Table("rooms").Order("rooms.id asc").Order("e.time_start asc").Select("rooms.*, e.time_start, e.time_end, e.allow_together").Joins("LEFT JOIN events AS e ON e.room_id = rooms.id").Rows()
//defer rows.Close()
//if err != nil {
//dbErrorLog(err)
//}
//for rows.Next() {
//seTime := StartEndTime{}
//var allowWith bool
//r := Room{}
//rows.Scan(&r.ID, &r.Place, &r.Date, &r.TimeStart, &r.TimeEnd, &seTime.TimeStart, &seTime.TimeEnd, &allowWith, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)
//// format
//r.Date = r.Date[:10]
//if len(rooms) == 0 || rooms[len(rooms)-1].ID != r.ID {
//if seTime.TimeStart.IsZero() && seTime.TimeEnd.IsZero() {
//availableTime = append(availableTime, StartEndTime{
//TimeStart: r.TimeStart,
//TimeEnd:   r.TimeEnd,
//})
//}
//rooms = append(rooms, r)
//}
//if seTime.TimeStart.IsZero() && seTime.TimeEnd.IsZero() {
//r = rooms[len(rooms)-1]
//r.calcAvailableTime(seTime, allowWith)
//rooms[len(rooms)-1] = r
//}
//}
//return rooms, nil
//}

func (room *Room) InTime(targetStartTime, targetEndTime time.Time) bool {
	for _, v := range room.calcAvailableTime() {
		roomStart := v.TimeStart
		roomEnd := v.TimeEnd
		if (roomStart.Equal(targetStartTime) || roomStart.Before(targetStartTime)) && (roomEnd.Equal(targetEndTime) || roomEnd.After(targetEndTime)) {
			fmt.Println(v)
			return true
		}
	}
	return false
}

// TODO return error
func (r *Room) calcAvailableTime() []StartEndTime {
	// TODO sort events by TimeStart
	availableTime := []StartEndTime{}
	availableTime = append(availableTime, StartEndTime{
		TimeStart: r.TimeStart,
		TimeEnd:   r.TimeEnd,
	})
	for _, event := range r.Events {
		if event.AllowTogether {
			continue
		}
		avleTimes := make([]StartEndTime, 2)
		i := len(availableTime) - 1
		avleTimes[0] = StartEndTime{
			TimeStart: availableTime[i].TimeStart,
			TimeEnd:   event.TimeStart,
		}
		avleTimes[1] = StartEndTime{
			TimeStart: event.TimeEnd,
			TimeEnd:   r.TimeEnd,
		}
		// delete last
		availableTime = availableTime[:len(availableTime)-1]
		for _, v := range avleTimes {
			if v.TimeStart != v.TimeEnd {
				availableTime = append(availableTime, v)
			}
		}
	}
	return availableTime
}

// isTimeContext 開始時間が終了時間より前か見る
func (r *Room) isTimeContext() bool {
	return r.TimeStart.Before(r.TimeEnd)
}

var traQCalendarID string = os.Getenv("TRAQ_CALENDARID")

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetEvents() ([]Room, error) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List(traQCalendarID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(100).OrderBy("startTime").Do()
	if err != nil {
		log.Printf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	rooms := []Room{}
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			room := Room{
				Place: item.Location,
				//Date:  date[:10],
				//TimeStart: item.Start.DateTime[11:16],
				//TimeEnd:   item.End.DateTime[11:16],
			}
			if err := DB.Set("gorm:insert_option", "ON DUPLICATE KEY UPDATE place=place").Create(&room).Error; err != nil {
				return nil, err
			}
			// 被りはIDが0で返ってくるらしい
			if room.ID != uuid.Nil {
				rooms = append(rooms, room)
			}

		}
	}
	return rooms, nil
}
