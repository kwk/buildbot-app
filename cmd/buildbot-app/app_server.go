package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
)

// AppServer stores all objects needed to talk to github, handle HTTP requests
// and talk to buildbot or receive
type AppServer struct {
	// This is where we
	GithubEventHandler *githubevents.EventHandler
	Mux                *http.ServeMux

	privateKeyFilePath  string
	appID               int64
	bindAddress         string
	githubWebhookSecret string

	buildbotMaster      string
	buildbotTryUser     string
	buildbotTryPassword string
}

// NewAppServer returns a new app server
func NewAppServer() (*AppServer, error) {
	appId, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse APP_ID: %w", err)
	}
	githubWebhookSecret := os.Getenv("APP_GITHUB_WEBHOOK_SECRET")
	return &AppServer{
		Mux:                 http.NewServeMux(),
		GithubEventHandler:  githubevents.New(githubWebhookSecret),
		privateKeyFilePath:  os.Getenv("APP_PRIVATE_KEY_FILE"),
		appID:               appId,
		bindAddress:         os.Getenv("APP_SERVER_BIND_ADDRESS"),
		githubWebhookSecret: githubWebhookSecret,
		buildbotMaster:      os.Getenv("BUILDBOT_MASTER"),
		buildbotTryUser:     os.Getenv("BUILDBOT_TRY_USER"),
		buildbotTryPassword: os.Getenv("BUILDBOT_TRY_PASSWORD"),
	}, nil
}

// NewGithubClient takes an installation ID and creates a
// github client targeting that very site.
func (srv *AppServer) NewGithubClient(appInstallationID int64) (*github.Client, error) {
	// Wrap the shared transport for use with the integration ID authenticating with installation ID 99.
	transport, err := ghinstallation.NewKeyFromFile(
		http.DefaultTransport,
		srv.appID,
		appInstallationID,
		srv.privateKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new key from file: %w", err)
	}
	// Use installation transport with github.com/google/go-github
	return github.NewClient(&http.Client{Transport: transport}), nil
}

// RunTryBot runs the "buildbot try" command against the configure buildbot
// master with a try-bot username and password. The command output is written to
// the logs.
func (srv *AppServer) RunTryBot(responsibleGithubLogin string, githubRepoOwner string, githubRepoName string, properties ...string) (string, error) {
	// In order to be able to run "buildbot try" from outside a git repository
	// we have to pass in an empty dummy diff file.
	dummyDiffFile, err := os.CreateTemp("/tmp", "dummy.*.diff")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dummyDiffFile.Name())

	args := append([]string{
		"try",
		fmt.Sprintf("--master=%s", srv.buildbotMaster),
		fmt.Sprintf("--builder=%s", "delegationBuilder"),
		fmt.Sprintf("--username=%s", srv.buildbotTryUser),
		fmt.Sprintf("--passwd=%s", srv.buildbotTryPassword),
		fmt.Sprintf("--diff=%s", dummyDiffFile.Name()),
		"--connect=pb",
		"--vc=git",
		fmt.Sprintf("--who=%s", responsibleGithubLogin),
		fmt.Sprintf("--repository=%s/%s", githubRepoOwner, githubRepoName),
	}, properties...)

	cmd := exec.Command("buildbot", args...)
	// TODO(kwk): Remove this command log to not show any password.
	// log.Printf("Running command: %s", cmd.String())
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ListenAndServer runs the app server's HTTP interface
func (srv *AppServer) ListenAndServe() {
	log.Printf("Listing for requests at http://%s\n", srv.bindAddress)
	log.Fatal(http.ListenAndServe(srv.bindAddress, srv.Mux))
}
