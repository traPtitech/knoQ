package utils

import (
	"github.com/cheekybits/genny/generic"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/presentation"
)

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "Type=uuid.UUID"

type Type generic.Type

func TypeIn(element Type, set []Type) bool {
	for _, v := range set {
		if element == v {
			// exist!
			return true
		}
	}
	return false
}

func ConvSchedule(src presentation.EventScheduleStatusReq) (dst domain.ScheduleStatus) {
	switch src.Schedule {
	case "pending":
		return domain.Pending
	case "attedance":
		return domain.Attedance
	case "absent":
		return domain.Absent
	}

	return 0
}
