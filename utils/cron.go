package utils

import (
	"fmt"
	repo "room/repository"
)

func initPostEventToTraQ(eventRepo repo.EventRepository) func() {
	job := func() {
		events, _ := eventRepo.GetAllEvents(nil, nil)
		fmt.Println(events[0].Name)
	}

	return job
}
