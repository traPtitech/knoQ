package domain

import (
	"reflect"
	"testing"
	"time"

	"github.com/traPtitech/knoQ/utils"
)

func TestRoom_CalcAvailableTime(t *testing.T) {
	now := time.Now()
	type fields struct {
		TimeStart time.Time
		TimeEnd   time.Time
		Events    []Event
	}
	tests := []struct {
		name          string
		fields        fields
		allowTogether bool
		want          []utils.TimeSpan
	}{
		{
			name: "simple",
			fields: fields{
				TimeStart: now,
				TimeEnd:   now.Add(10 * time.Hour),
				Events: []Event{
					{
						TimeStart:     now.Add(1 * time.Hour),
						TimeEnd:       now.Add(2 * time.Hour),
						AllowTogether: false,
					},
				},
			},
			want: []utils.TimeSpan{
				{
					Start: now,
					End:   now.Add(1 * time.Hour),
				},
				{
					Start: now.Add(2 * time.Hour),
					End:   now.Add(10 * time.Hour),
				},
			},
			allowTogether: true,
		},
		{
			name: "edge",
			fields: fields{
				TimeStart: now,
				TimeEnd:   now.Add(10 * time.Hour),
				Events: []Event{
					{
						TimeStart:     now,
						TimeEnd:       now.Add(10 * time.Hour),
						AllowTogether: false,
					},
				},
			},
			want:          []utils.TimeSpan{},
			allowTogether: true,
		},
		{
			name: "Intersection",
			fields: fields{
				TimeStart: now,
				TimeEnd:   now.Add(10 * time.Hour),
				Events: []Event{
					{
						TimeStart:     now,
						TimeEnd:       now.Add(3 * time.Hour),
						AllowTogether: true,
					},
					{
						TimeStart:     now.Add(2 * time.Hour),
						TimeEnd:       now.Add(4 * time.Hour),
						AllowTogether: true,
					},
				},
			},
			want: []utils.TimeSpan{
				{
					Start: now.Add(4 * time.Hour),
					End:   now.Add(10 * time.Hour),
				},
			},
			allowTogether: false,
		},
		{
			name: "Independent error",
			fields: fields{
				TimeStart: now,
				TimeEnd:   now.Add(10 * time.Hour),
				Events: []Event{
					{
						TimeStart: now.Add(4 * time.Hour),
						TimeEnd:   now.Add(12 * time.Hour),
					},
				},
			},
			want: []utils.TimeSpan{
				{
					Start: now,
					End:   now.Add(4 * time.Hour),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Room{
				TimeStart: tt.fields.TimeStart,
				TimeEnd:   tt.fields.TimeEnd,
				Events:    tt.fields.Events,
			}
			got := r.CalcAvailableTime(tt.allowTogether)
			if !reflect.DeepEqual(got, tt.want) {
				if !(len(got) == 0 && len(tt.want) == 0) {
					t.Errorf("r.CalcAvailableTime() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
