package utils

import "time"

func StrToTime(s string) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		t, err = time.Parse("15:04:05", s)
		if err != nil {
			return time.Time{}, err
		}
	}
	return t, nil
}
