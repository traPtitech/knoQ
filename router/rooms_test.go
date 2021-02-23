package router

import (
	"reflect"
	"testing"
	"time"

	repo "github.com/traPtitech/knoQ/repository"
)

func Test_filterSameRooms(t *testing.T) {
	type args struct {
		currentRooms []*repo.Room
		targetRooms  []*repo.Room
	}
	baseRoom := repo.Room{
		Place:     "sample place",
		Public:    true,
		TimeStart: time.Now(),
		TimeEnd:   time.Now().Add(3 * time.Hour),
	}
	room1 := baseRoom
	room1.Place = "room1"
	room2 := baseRoom
	room2.Place = "room2"
	room3 := baseRoom
	room3.Place = "room3"

	tests := []struct {
		name string
		args args
		want []*repo.Room
	}{
		{
			name: "nil",
			args: args{
				currentRooms: []*repo.Room{
					&baseRoom,
				},
				targetRooms: []*repo.Room{
					&baseRoom,
				},
			},
			want: []*repo.Room{},
		},
		{
			name: "one ",
			args: args{
				currentRooms: []*repo.Room{
					&room1, &room2,
				},
				targetRooms: []*repo.Room{
					&room1, &room2, &room3,
				},
			},
			want: []*repo.Room{
				&room3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterSameRooms(tt.args.currentRooms, tt.args.targetRooms); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterSameRooms() = %v, want %v", got, tt.want)
			}
		})
	}
}
