package repository

import (
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	traQutils "github.com/traPtitech/traQ/utils"
)

func TestGormRepository_CreateEvent(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	group := mustMakeGroup(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)
	room := mustMakeRoom(t, repo, traQutils.RandAlphabetAndNumberString(10))

	params := WriteEventParams{
		Name:      traQutils.RandAlphabetAndNumberString(20),
		GroupID:   group.ID,
		RoomID:    room.ID,
		TimeStart: time.Now(),
		TimeEnd:   time.Now().Add(1 * time.Minute),
		CreatedBy: user.ID,
	}

	if event, err := repo.CreateEvent(params); assert.NoError(t, err) {
		assert.NotNil(t, event)
	}

}

func TestGormRepository_DeleteTagInEvent(t *testing.T) {
	t.Parallel()
	repo, _, _, user := setupGormRepoWithUser(t, common)
	event, _, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)

	t.Run("delete unlocked tag when deleteLocked == false", func(t *testing.T) {
		t.Parallel()
		tag := mustMakeTag(t, repo, traQutils.RandAlphabetAndNumberString(10))
		err := repo.AddTagToEvent(event.ID, tag.ID, false)
		require.NoError(t, err)

		err = repo.DeleteTagInEvent(event.ID, tag.ID, false)
		assert.NoError(t, err)

	})

	t.Run("delete locked tag when deleteLocked == false", func(t *testing.T) {
		t.Parallel()
		tag := mustMakeTag(t, repo, traQutils.RandAlphabetAndNumberString(10))
		err := repo.AddTagToEvent(event.ID, tag.ID, true)
		require.NoError(t, err)

		err = repo.DeleteTagInEvent(event.ID, tag.ID, false)
		assert.Error(t, err)
	})

	t.Run("delete locked tag when deleteLocked == true", func(t *testing.T) {
		t.Parallel()
		tag := mustMakeTag(t, repo, traQutils.RandAlphabetAndNumberString(10))
		err := repo.AddTagToEvent(event.ID, tag.ID, true)
		require.NoError(t, err)

		err = repo.DeleteTagInEvent(event.ID, tag.ID, true)
		assert.NoError(t, err)
	})

}

func TestGormRepository_UpdateEvent(t *testing.T) {
	repo, _, _, user := setupGormRepoWithUser(t, common)
	e, _, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)

	params := WriteEventParams{
		Name:      e.Name,
		GroupID:   e.GroupID,
		RoomID:    e.RoomID,
		TimeStart: time.Now(),
		TimeEnd:   time.Now().Add(10 * time.Minute),
		CreatedBy: e.CreatedBy,
	}

	if event, err := repo.UpdateEvent(e.ID, params); assert.NoError(t, err) {
		assert.Equal(t, e.ID, event.ID)
		assert.Equal(t, params.TimeEnd, event.TimeEnd)
	}
}

func TestGormRepository_GetEventsByGroupIDs(t *testing.T) {
	repo, _, _, user := setupGormRepoWithUser(t, common)
	_, g1, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)
	_, g2, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)

	if events, err := repo.GetEventsByGroupIDs([]uuid.UUID{g1.ID, g2.ID}); assert.NoError(t, err) {
		assert.Equal(t, 2, len(events))
	}
}

func TestGormRepository_GetEvent(t *testing.T) {
	repo, _, _, user := setupGormRepoWithUser(t, common)
	e, _, _ := mustMakeEvent(t, repo, traQutils.RandAlphabetAndNumberString(10), user.ID)

	t.Run("Get an existing event", func(t *testing.T) {
		if event, err := repo.GetEvent(e.ID); assert.NoError(t, err) {
			assert.Equal(t, e.ID, event.ID)
		}
	})
}

func ExampleEvent_ICal() {
	t, _ := time.Parse("20060102T150405Z", "20060102T150405Z")
	e := &Event{
		ID:          uuid.Nil,
		Name:        "sample",
		Description: "aiueo\naa\n[aa](aa)",
		Room: Room{
			Place: "S516",
		},
		TimeStart: t,
		TimeEnd:   t.Add(3 * time.Hour),
		CreatedBy: uuid.Nil,
		Model: Model{
			CreatedAt: t,
			UpdatedAt: t,
		},
	}
	c := ical.New()
	ical.NewEvent()
	tz := ical.NewTimezone()
	tz.AddProperty("TZID", "Asia/Tokyo")
	std := ical.NewStandard()
	std.AddProperty("TZOFFSETFROM", "+9000")
	std.AddProperty("TZOFFSETTO", "+9000")
	std.AddProperty("TZNAME", "JST")
	std.AddProperty("DTSTART", "19700101T000000")
	tz.AddEntry(std)
	c.AddEntry(tz)

	// event
	vevent := e.ICal()
	// override for test
	vevent.AddProperty("dtstamp", t.Format("20060102T150405Z"))
	c.AddEntry(vevent)

	ical.NewEncoder(os.Stdout).Encode(c)

	// Output:
	// BEGIN:VCALENDAR
	// VERSION:2.0
	// PRODID:github.com/lestrrat-go/ical
	// BEGIN:VTIMEZONE
	// TZID:Asia/Tokyo
	// BEGIN:STANDARD
	// DTSTART:19700101T000000
	// TZNAME:JST
	// TZOFFSETFROM:+9000
	// TZOFFSETTO:+9000
	// END:STANDARD
	// END:VTIMEZONE
	// BEGIN:VEVENT
	// CREATED:20060102T150405Z
	// DESCRIPTION:aiueo\naa\n[aa](aa)
	// DTEND:20060102T180405Z
	// DTSTAMP:20060102T150405Z
	// DTSTART:20060102T150405Z
	// LAST-MODIFIED:20060102T150405Z
	// LOCATION:S516
	// ORGANIZER:00000000-0000-0000-0000-000000000000
	// SUMMARY:sample
	// UID:00000000-0000-0000-0000-000000000000
	// END:VEVENT
	// END:VCALENDAR
}
