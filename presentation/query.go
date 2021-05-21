package presentation

import (
	"net/url"
	"time"
)

// getTimeRange ?dateBegin=2020-03-27T00:00:00Z
func GetTiemRange(values url.Values) (start time.Time, end time.Time, err error) {
	if values.Get("dateBegin") != "" {
		start, err = time.Parse(time.RFC3339, values.Get("dateBegin"))
		if err != nil {
			return
		}
	}
	if values.Get("dateEnd") != "" {
		end, err = time.Parse(time.RFC3339, values.Get("dateEnd"))
		if err != nil {
			return
		}
	}
	return
}

type UserRelation int

const (
	RelationBelongs = iota
	RelationAdmins  = iota
)

func GetUserRelationQuery(values url.Values) UserRelation {
	relation := values.Get("relation")
	switch relation {
	case "belongs":
		return RelationBelongs
	case "admins":
		return RelationAdmins
	}

	return RelationBelongs
}
