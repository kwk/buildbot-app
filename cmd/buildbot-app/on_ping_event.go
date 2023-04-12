package main

import (
	"log"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
)

func (srv *AppServer) OnPingEventAny() githubevents.PingEventHandleFunc {
	return func(deliveryID, eventName string, event *github.PingEvent) error {
		log.Println("Ping event")
		return nil
	}
}
