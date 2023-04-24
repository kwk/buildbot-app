package main

import (
	"log"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
)

func (srv *AppServer) OnCheckRunEventRequestAction() githubevents.CheckRunEventHandleFunc {
	return func(deliveryID string, eventName string, event *github.CheckRunEvent) error {
		log.Printf("NOT IMPLEMENTED: OnCheckRunEventRequestAction with this requested action identifier: %s\n", event.RequestedAction.Identifier)
		return nil
	}
}
