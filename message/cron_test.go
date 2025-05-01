package message

import (
	"reflect"
	"testing"
	"time"

	"github.com/traPtitech/knoQ/domain"
)

func Test_makeRoomAvailableByTimeTable(t *testing.T) {
	today := time.Now()
	tt := []timeTable{
		{":sunny:", setTimeFromString(today, "00:00:00"), false},
		{"1-2", setTimeFromString(today, "08:50:00"), true},
		{"3-4", setTimeFromString(today, "10:45:00"), true},
		{"昼", setTimeFromString(today, "12:25:00"), true},
		{"5-6", setTimeFromString(today, "13:30:00"), true},
		{"7-8", setTimeFromString(today, "15:25:00"), true},
		{"9-10", setTimeFromString(today, "17:15:00"), true},
		{":crescent_moon:", setTimeFromString(today, "18:55:00"), false},
	}

	stampAvailable := ":white_check_mark:"
	stampNotAvailable := ":regional_indicator_null:"
	tests := map[string]struct {
		room []*domain.Room
		want []map[string]string
	}{
		"1-10": {
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(today, "08:50:00"),
					TimeEnd:   setTimeFromString(today, "18:55:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNotAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampNotAvailable},
			},
		},
		"cut unverified": {
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(today, "08:50:00"),
					TimeEnd:   setTimeFromString(today, "18:55:00"),
				},
				{
					Place:     "unverified",
					Verified:  false,
					TimeStart: setTimeFromString(today, "08:50:00"),
					TimeEnd:   setTimeFromString(today, "18:55:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNotAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampNotAvailable},
			},
		},
		"24/7": {
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(today, "00:00:00"),
					TimeEnd:   setTimeFromString(today, "23:59:59"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
			},
		},
		"5-10 and 3-10 except Lunch": {
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(today, "12:25:00"),
					TimeEnd:   setTimeFromString(today, "18:55:00"),
				},
				{
					Place:     "traP-002",
					Verified:  true,
					TimeStart: setTimeFromString(today, "10:45:00"),
					TimeEnd:   setTimeFromString(today, "12:25:00"),
				},
				{
					Place:     "traP-002",
					Verified:  true,
					TimeStart: setTimeFromString(today, "13:30:00"),
					TimeEnd:   setTimeFromString(today, "18:55:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNotAvailable, "traP-002": stampNotAvailable},
				{"traP-001": stampNotAvailable, "traP-002": stampNotAvailable},
				{"traP-001": stampNotAvailable, "traP-002": stampAvailable},
				{"traP-001": stampAvailable, "traP-002": stampNotAvailable},
				{"traP-001": stampAvailable, "traP-002": stampAvailable},
				{"traP-001": stampAvailable, "traP-002": stampAvailable},
				{"traP-001": stampAvailable, "traP-002": stampAvailable},
				{"traP-001": stampNotAvailable, "traP-002": stampNotAvailable},
			},
		},
		"10:00-18:00": {
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(today, "10:00:00"),
					TimeEnd:   setTimeFromString(today, "18:00:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNotAvailable},
				{"traP-001": "10:00 -"},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": stampAvailable},
				{"traP-001": "- 18:00"},
				{"traP-001": stampNotAvailable},
			},
		},
		"15:00-15:01": {
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(today, "15:00:00"),
					TimeEnd:   setTimeFromString(today, "15:01:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNotAvailable},
				{"traP-001": stampNotAvailable},
				{"traP-001": stampNotAvailable},
				{"traP-001": stampNotAvailable},
				{"traP-001": "15:00 - 15:01"},
				{"traP-001": stampNotAvailable},
				{"traP-001": stampNotAvailable},
				{"traP-001": stampNotAvailable},
			},
		},
	}
	for name, te := range tests {
		t.Run(name, func(t *testing.T) {
			got := makeRoomAvailableByTimeTable(te.room, tt, today)

			if !reflect.DeepEqual(got, te.want) {
				if !(len(got) == 0 && len(te.want) == 0) {
					t.Errorf("\ngot : %v\nwant: %v", got, te.want)
				}
			}
		})
	}
}
