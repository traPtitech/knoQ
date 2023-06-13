package utils

import (
	"github.com/traPtitech/knoQ/domain"
	"reflect"
	"time"

	"testing"
)

func Test_makeRoomAvailableByTimeTable(t *testing.T) {

	date := time.Now()

	tt := []timeTable{
		{":sunny:", setTimeFromString(date, "00:00:00"), false},
		{"1-2", setTimeFromString(date, "08:50:00"), true},
		{"3-4", setTimeFromString(date, "10:45:00"), true},
		{"昼", setTimeFromString(date, "12:25:00"), true},
		{"5-6", setTimeFromString(date, "13:45:00"), true},
		{"7-8", setTimeFromString(date, "15:40:00"), true},
		{"9-10", setTimeFromString(date, "17:30:00"), true},
		{":crescent_moon:", setTimeFromString(date, "19:10:00"), false},
	}

	stampOk := ":white_check_mark:"
	stampNo := ":regional_indicator_null:"
	tests := []struct {
		name string
		room []*domain.Room
		want []map[string]string
	}{
		{
			name: "1-10",
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(date, "08:50:00"),
					TimeEnd:   setTimeFromString(date, "19:10:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNo},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampNo},
			},
		},
		{
			name: "cut unverified",
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(date, "08:50:00"),
					TimeEnd:   setTimeFromString(date, "19:10:00"),
				},
				{
					Place:     "unverified",
					Verified:  false,
					TimeStart: setTimeFromString(date, "08:50:00"),
					TimeEnd:   setTimeFromString(date, "19:10:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNo},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampNo},
			},
		},
		{
			name: "24/7",
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(date, "00:00:00"),
					TimeEnd:   setTimeFromString(date, "23:59:59"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
			},
		},
		{
			name: "5-10 and 3-10 except Lunch",
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(date, "12:25:00"),
					TimeEnd:   setTimeFromString(date, "19:10:00"),
				},
				{
					Place:     "traP-002",
					Verified:  true,
					TimeStart: setTimeFromString(date, "10:45:00"),
					TimeEnd:   setTimeFromString(date, "12:25:00"),
				},
				{
					Place:     "traP-002",
					Verified:  true,
					TimeStart: setTimeFromString(date, "13:45:00"),
					TimeEnd:   setTimeFromString(date, "19:10:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNo, "traP-002": stampNo},
				{"traP-001": stampNo, "traP-002": stampNo},
				{"traP-001": stampNo, "traP-002": stampOk},
				{"traP-001": stampOk, "traP-002": stampNo},
				{"traP-001": stampOk, "traP-002": stampOk},
				{"traP-001": stampOk, "traP-002": stampOk},
				{"traP-001": stampOk, "traP-002": stampOk},
				{"traP-001": stampNo, "traP-002": stampNo},
			},
		},
		{
			name: "10:00-18:00",
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(date, "10:00:00"),
					TimeEnd:   setTimeFromString(date, "18:00:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNo},
				{"traP-001": "10:00 -"},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": stampOk},
				{"traP-001": "- 18:00"},
				{"traP-001": stampNo},
			},
		},
		{
			name: "15:00-15:01",
			room: []*domain.Room{
				{
					Place:     "traP-001",
					Verified:  true,
					TimeStart: setTimeFromString(date, "15:00:00"),
					TimeEnd:   setTimeFromString(date, "15:01:00"),
				},
			},
			want: []map[string]string{
				{"traP-001": stampNo},
				{"traP-001": stampNo},
				{"traP-001": stampNo},
				{"traP-001": stampNo},
				{"traP-001": "15:00 - 15:01"},
				{"traP-001": stampNo},
				{"traP-001": stampNo},
				{"traP-001": stampNo},
			},
		},
	}
	for _, te := range tests {
		t.Run(te.name, func(t *testing.T) {
			got := makeRoomAvailableByTimeTable(te.room, tt, time.Now())

			if !reflect.DeepEqual(got, te.want) {
				if !(len(got) == 0 && len(te.want) == 0) {
					t.Errorf("\ngot : %v\nwant: %v", got, te.want)
				}
			}
		})
	}
}
