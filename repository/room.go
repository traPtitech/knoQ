package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"room/utils"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func FindRoomsByTime(begin, end string) ([]Room, error) {
	rooms := []Room{}
	cmd := DB
	if begin != "" {
		cmd = cmd.Where("date >= ?", begin)
	}
	if end != "" {
		cmd = cmd.Where("date <= ?", end)
	}

	if err := cmd.Order("date asc").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

// AddRelation add room by RoomID
func (room *Room) AddRelation(roomID int) error {
	room.ID = roomID
	if err := DB.First(&room, room.ID).Error; err != nil {
		return err
	}
	return nil
}

func (room *Room) InTime(targetTime time.Time) bool {
	roomStart, _ := utils.StrToTime(room.TimeStart)
	roomEnd, _ := utils.StrToTime(room.TimeEnd)
	if (roomStart.Equal(targetTime) || roomStart.Before(targetTime)) && (roomEnd.Equal(targetTime) || roomEnd.After(targetTime)) {
		return true
	}
	return false
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
				Place:     item.Location,
				Date:      date[:10],
				TimeStart: item.Start.DateTime[11:16],
				TimeEnd:   item.End.DateTime[11:16],
			}
			if err := DB.Set("gorm:insert_option", "ON DUPLICATE KEY UPDATE place=place").Create(&room).Error; err != nil {
				return nil, err
			}
			// 被りはIDが0で返ってくるらしい
			if room.ID != 0 {
				rooms = append(rooms, room)
			}

		}
	}
	return rooms, nil
}
