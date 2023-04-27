package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/kwk/buildbot-app/cmd/buildbot-app/buildbot_http_status_push"
)

func (srv *AppServer) HandleBuildBotStatusHook() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("/buildbot-status-hook called\n")

		// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
		// response body. A request body larger than that will now result in
		// Decode() returning a "http: request body too large" error.
		req.Body = http.MaxBytesReader(w, req.Body, 1048576)

		// Decode JSON payload into Go structure
		var buildStatus buildbot_http_status_push.Data

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
		appInstallId, err := strconv.ParseInt(buildStatus.Properties.GithubAppInstallationID[0], 10, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse buildStatus.Properties.GithubAppInstallationID[0] as int64: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gh, err := srv.NewGithubClient(appInstallId)
		if err != nil {
			err = fmt.Errorf("error creating github client: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		checkRunID, err := strconv.ParseInt(buildStatus.Properties.GithubCheckRunID[0], 10, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse buildStatus.Properties.GithubCheckRunID[0] as int64: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		checkRun, _, err := gh.Checks.GetCheckRun(req.Context(), buildStatus.Properties.GithubPullRequestRepoOwner[0], buildStatus.Properties.GithubPullRequestRepoName[0], checkRunID)
		if err != nil {
			err = fmt.Errorf("error getting check run: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		conclusion := CheckRunStateFromBuildbotResult(buildStatus.Results)

		now := time.Now()
		newStateString := fmt.Sprintf("[Builder: %s]: %s ([log](%s))", buildStatus.Builder.Name, buildStatus.StateString, buildStatus.URL)
		if checkRun.Output != nil && checkRun.Output.Summary != nil {
			newStateString = strings.Join([]string{*checkRun.Output.Summary, WrapMsgWithTimePrefix(newStateString, now)}, "\n")
		}

		_, _, err = gh.Checks.UpdateCheckRun(req.Context(), buildStatus.Properties.GithubPullRequestRepoOwner[0], buildStatus.Properties.GithubPullRequestRepoName[0], checkRunID, github.UpdateCheckRunOptions{
			Name:       *checkRun.Name,
			Status:     github.String(string(CheckRunStateCompleted)),
			Conclusion: github.String(string(conclusion)),
			DetailsURL: github.String(buildStatus.URL),
			Output: &github.CheckRunOutput{
				Title:   github.String("Buildbot Status Log"),
				Summary: github.String(newStateString),
				Text:    github.String(fmt.Sprintf("[Buildbot Build Page](%s)", buildStatus.URL)),
			},
			Actions: []*github.CheckRunAction{
				{
					Label:       "Make check required",
					Description: "Make check required to pass",
					Identifier:  "MakeMandatory",
				},
				{
					Label:       "Make check optional",
					Description: "This check is optional",
					Identifier:  "MakeOptional",
				},
				{
					Label:       "Rerun check",
					Description: "Reruns the check",
					Identifier:  "ReRunCheck",
				},
			},
		})
		if err != nil {
			err := fmt.Errorf("failed to update try bot check run: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("updated github check run: %s\n", *checkRun.Name)
		log.Printf("check run details: %s\n", buildStatus.URL)

		// Update the build log comment
		//-------------------------------------
		buildLogCommentID, err := strconv.ParseInt(buildStatus.Properties.GithubBuildLogCommentID[0], 10, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse buildStatus.Properties.GithubBuildLogCommentID[0] as int64: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buildLogComment, _, err := gh.Issues.GetComment(req.Context(), buildStatus.Properties.GithubPullRequestRepoOwner[0], buildStatus.Properties.GithubPullRequestRepoName[0], buildLogCommentID)
		if err != nil {
			err := fmt.Errorf("failed to get build log comment: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buildLogCommentBody := buildLogComment.Body
		newBuildLogComment := buildLogComment
		newBuildLogComment.Body = github.String(fmt.Sprintf(`%s<br/><strong>%s</strong> <i>[Builder: %s]</i> %s (<a href="%s">log</a>)`, *buildLogCommentBody, now.Format(time.RFC1123Z), buildStatus.Builder.Name, buildStatus.StateString, buildStatus.URL))
		_, _, err = gh.Issues.EditComment(req.Context(), buildStatus.Properties.GithubPullRequestRepoOwner[0], buildStatus.Properties.GithubPullRequestRepoName[0], buildLogCommentID, newBuildLogComment)
		if err != nil {
			err := fmt.Errorf("failed to edit build log comment: %w", err)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		io.WriteString(w, "thank you for calling back to the buildbot-app")
	}
}
