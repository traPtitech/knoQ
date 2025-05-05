package utils

import (
	"testing"
	"time"
)

func TestSubtractTimeSpan(t *testing.T) {
	layout := "15:04"

	parse := func(s string) time.Time {
		tm, _ := time.Parse(layout, s)
		return tm
	}

	tests := []struct {
		base TimeSpan
		sub  TimeSpan
		want []TimeSpan
	}{
		// 完全に重ならない (前)
		{
			base: TimeSpan{parse("10:00"), parse("12:00")},
			sub:  TimeSpan{parse("08:00"), parse("09:00")},
			want: []TimeSpan{{parse("10:00"), parse("12:00")}},
		},
		// 完全に重ならない (後)
		{
			base: TimeSpan{parse("10:00"), parse("12:00")},
			sub:  TimeSpan{parse("13:00"), parse("14:00")},
			want: []TimeSpan{{parse("10:00"), parse("12:00")}},
		},
		// base を完全に覆う sub
		{
			base: TimeSpan{parse("10:00"), parse("12:00")},
			sub:  TimeSpan{parse("09:00"), parse("13:00")},
			want: []TimeSpan{},
		},
		// sub が base の一部を切る (前半)
		{
			base: TimeSpan{parse("10:00"), parse("12:00")},
			sub:  TimeSpan{parse("10:30"), parse("11:00")},
			want: []TimeSpan{
				{parse("10:00"), parse("10:30")},
				{parse("11:00"), parse("12:00")},
			},
		},
		// sub が base の前半に重なる
		{
			base: TimeSpan{parse("10:00"), parse("12:00")},
			sub:  TimeSpan{parse("09:30"), parse("11:00")},
			want: []TimeSpan{
				{parse("11:00"), parse("12:00")},
			},
		},
		// sub が base の後半に重なる
		{
			base: TimeSpan{parse("10:00"), parse("12:00")},
			sub:  TimeSpan{parse("11:00"), parse("13:00")},
			want: []TimeSpan{
				{parse("10:00"), parse("11:00")},
			},
		},
		// sub と base が同じ
		{
			base: TimeSpan{parse("10:00"), parse("12:00")},
			sub:  TimeSpan{parse("10:00"), parse("12:00")},
			want: []TimeSpan{},
		},
	}

	for i, tt := range tests {
		got := SubtractTimeSpan(tt.base, tt.sub)
		if !equalTimeSpans(got, tt.want) {
			t.Errorf("case %d: got %v, want %v", i, got, tt.want)
		}
	}
}

func equalTimeSpans(a, b []TimeSpan) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Start.Equal(b[i].Start) || !a[i].End.Equal(b[i].End) {
			return false
		}
	}
	return true
}
