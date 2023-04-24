package main

import (
	"log"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
)

func (srv *AppServer) OnCheckRunEventReRequested() githubevents.CheckRunEventHandleFunc {
	return func(deliveryID string, eventName string, event *github.CheckRunEvent) error {
		log.Printf("NOT IMPLEMENTED: OnCheckRunEventReRequested not implemented yet")
		return nil
	}
}
