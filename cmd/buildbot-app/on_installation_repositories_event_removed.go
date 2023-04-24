package main

import (
	"log"
	"strings"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
)

func (srv *AppServer) OnInstallationRepositoriesEventRemoved() githubevents.InstallationRepositoriesEventHandleFunc {
	return func(deliveryID string, eventName string, event *github.InstallationRepositoriesEvent) error {
		html_urls := make([]string, len(event.RepositoriesRemoved))
		for i := 0; i < len(event.RepositoriesRemoved); i++ {
			html_urls[i] = *event.RepositoriesRemoved[i].HTMLURL
		}
		log.Printf("%s uninstalled app from %s", *event.Sender.Login, strings.Join(html_urls, ","))
		return nil
	}
}
