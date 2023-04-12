package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/go-github/v50/github"
)

// BuildStatus is what we expect buildbot to report back to the Github App, a
// very simple struct for now.
type BuildStatus struct {
	GithubPullRequest
	BuildbotBuildStatus  string `json:"buildbot_build_status" binding:"required"`
	BuildbotBuildHTMLURL string `json:"buildbot_build_html_url" binding:"required"`
	BuildbotWorkerName   string `json:"buildbot_worker_name" binding:"required"`

	// GithubCheckRunId of the check run associated with this build.
	GithubCheckRunId int64 `json:"github_check_run_id" binding:"required"`
	// TODO(kwk): GithubAppInstallationID is really just a hack to be able to
	// create a github client for the specific application installation.
	GithubAppInstallationID int64 `json:"github_app_installation_id" binding:"required"`
}

func (srv *AppServer) HandleBuildBotHook() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("/buildbot-hook called\n")

		// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
		// response body. A request body larger than that will now result in
		// Decode() returning a "http: request body too large" error.
		req.Body = http.MaxBytesReader(w, req.Body, 1048576)

		// Decode JSON payload into Go structure
		var buildStatus BuildStatus

		// Setup the decoder and call the DisallowUnknownFields() method on it.
		// This will cause Decode() to return a "json: unknown field ..." error
		// if it encounters any extra unexpected fields in the JSON. Strictly
		// speaking, it returns an error for "keys which do not match any
		// non-ignored, exported fields in the destination".
		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()

		err := dec.Decode(&buildStatus)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("buildStatus: %+v\n", buildStatus)

		// Update the github check run associated with this build status
		// Create a github client based for this app's installation
		gh, err := srv.NewGithubClient(buildStatus.GithubAppInstallationID)
		if err != nil {
			err = fmt.Errorf("error creating github client: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		checkRun, _, err := gh.Checks.GetCheckRun(req.Context(), buildStatus.BaseRepoOwner, buildStatus.BaseRepoName, buildStatus.GithubCheckRunId)
		if err != nil {
			err = fmt.Errorf("error getting check run: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _, err = gh.Checks.UpdateCheckRun(req.Context(), buildStatus.BaseRepoOwner, buildStatus.BaseRepoName, buildStatus.GithubCheckRunId, github.UpdateCheckRunOptions{
			Name:       *checkRun.Name,
			Status:     github.String(string(CheckRunStateCompleted)),
			Conclusion: github.String(string(CheckRunConclusionSuccess)), // TODO(kwk): Take actual build status
			DetailsURL: github.String(buildStatus.BuildbotBuildHTMLURL),
			Output: &github.CheckRunOutput{
				Title:   github.String("My title"),
				Summary: github.String("My summary"),
				Text:    github.String(fmt.Sprintf("[Buildbot Build Page](%s)", buildStatus.BuildbotBuildHTMLURL)),
			},
		})
		if err != nil {
			err := fmt.Errorf("failed to update try bot check run: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("updated github check run: %s\n", *checkRun.Name)
		log.Printf("check run details: %s\n", buildStatus.BuildbotBuildHTMLURL)
		io.WriteString(w, "thank you for calling back to the buildbot-app")
	}
}
