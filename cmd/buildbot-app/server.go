package main

import (
	"github.com/google/go-github/v50/github"
)

// tag::server[]
// Server specifies the interface that we need to implement from the AppServer
// object in order to provide a decent mock in tests.
type Server interface {

	// NewGithubClient returns a new GitHub client object for the given
	// application ID.
	NewGithubClient(appInstallationID int64) (*github.Client, error)

	// RunTryBot runs a "buildbot try" command
	RunTryBot(responsibleGithubLogin string, githubRepoOwner string, githubRepoName string, properties ...string) (string, error)
}

// end::server[]
