package main

import (
	"log"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
)

func (srv *AppServer) OnInstallationEventCreated() githubevents.InstallationEventHandleFunc {
	return func(deliveryID string, eventName string, event *github.InstallationEvent) error {
		log.Printf("%s installed app to %s", *event.Sender.Login, *event.Installation.RepositoriesURL)
		return nil
	}
}
