package main

import (
	"log"
)

func main() {
	srv, err := NewAppServer()
	if err != nil {
		log.Fatalf("failed to load config: %+v", err)
	}

	// A simple test endpoint
	srv.Mux.HandleFunc("/foobar", srv.HandleFoobar())

	// When github wants to ping the app
	srv.GithubEventHandler.OnPingEventAny(srv.OnPingEventAny())

	// This is where we're going to handle /buildbot comments made on github
	srv.GithubEventHandler.OnIssueCommentEventAny(srv.OnIssueCommentEventAny())

	// When buildbot wants to talk to the Github App it can use this endpoint
	srv.Mux.HandleFunc("/buildbot-hook", srv.HandleBuildBotHook())

	// This gets called when you have a check run with an action and someone
	// clicks on the button in the github check run page.
	// TODO(kwk): Think if this could be useful.
	srv.GithubEventHandler.OnCheckRunEventRequestAction(srv.OnCheckRunEventRequestAction())

	// This is the entrypoint for Webhooks coming from Github
	// NOTE: Make sure to have setup the GithubEventHandler beforehand
	srv.Mux.HandleFunc("/github-hook", srv.HandleGithubHook())

	srv.ListenAndServe()
}
