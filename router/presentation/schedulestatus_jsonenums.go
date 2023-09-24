// Code generated by jsonenums -type=ScheduleStatus; DO NOT EDIT.

package presentation

import (
	"encoding/json"
	"fmt"
)

var (
	_ScheduleStatusNameToValue = map[string]ScheduleStatus{
		"pending":    Pending,
		"attendance": Attendance,
		"absent":     Absent,
	}

	_ScheduleStatusValueToName = map[ScheduleStatus]string{
		Pending:    "pending",
		Attendance: "attendance",
		Absent:     "absent",
	}
)

func init() {
	var v ScheduleStatus
	if _, ok := interface{}(v).(fmt.Stringer); ok {
		_ScheduleStatusNameToValue = map[string]ScheduleStatus{
			interface{}(Pending).(fmt.Stringer).String():    Pending,
			interface{}(Attendance).(fmt.Stringer).String(): Attendance,
			interface{}(Absent).(fmt.Stringer).String():     Absent,
		}
	}
}

// MarshalJSON is generated so ScheduleStatus satisfies json.Marshaler.
func (r ScheduleStatus) MarshalJSON() ([]byte, error) {
	if s, ok := interface{}(r).(fmt.Stringer); ok {
		return json.Marshal(s.String())
	}
	s, ok := _ScheduleStatusValueToName[r]
	if !ok {
		return nil, fmt.Errorf("invalid ScheduleStatus: %d", r)
	}
	return json.Marshal(s)
}

// UnmarshalJSON is generated so ScheduleStatus satisfies json.Unmarshaler.
func (r *ScheduleStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ScheduleStatus should be a string, got %s", data)
	}
	v, ok := _ScheduleStatusNameToValue[s]
	if !ok {
		return fmt.Errorf("invalid ScheduleStatus %q", s)
	}
	*r = v
	return nil
}