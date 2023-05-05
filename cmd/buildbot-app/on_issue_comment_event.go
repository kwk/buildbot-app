package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
	"github.com/kwk/buildbot-app/cmd/buildbot-app/command"
)

func OnIssueCommentEventAny(srv Server) githubevents.IssueCommentEventHandleFunc {
	return func(deliveryID string, eventName string, event *github.IssueCommentEvent) error {
		if event == nil {
			return nil
		}
		// Only run on PR comments and not on issue comments
		if !event.Issue.IsPullRequest() {
			return nil
		}
		// Only handle new comments or edited ones
		switch event.GetAction() {
		case "edited":
			break
		case "created":
			break
		default:
			return nil
		}
		comment := event.Comment
		if comment == nil {
			return nil
		}
		if comment.Body == nil {
			return nil
		}
		// Check if comment body can be parsed as a /buildbot
		if !command.StringIsCommand(*comment.Body) {
			return nil
		}
		log.Printf("/buildbot was used")

		// tag::thank_you[]
		// This comment will be used all over the place
		thankYouComment := fmt.Sprintf(
			`Thank you @%s for using the <a href="todo:link-to-documentation-here"><code>%s</code></a> command <a href="%s">here</a>! `,
			command.BuildbotCommand,
			*event.Comment.User.Login,
			*event.Comment.HTMLURL,
		)

		// end::thank_you[]

		cmd, err := command.FromString(*comment.Body)
		if err != nil {
			return fmt.Errorf("failed to parse command: %w", err)
			// TODO(kwk): Maybe tell the user that we didn't understand the request
		}
		cmd.CommentAuthor = *comment.User.Login

		// Create a github client based for this app's installation
		appInstallationID := *event.GetInstallation().ID
		gh, err := srv.NewGithubClient(appInstallationID)
		if err != nil {
			err = fmt.Errorf("error creating github client: %w", err)
			log.Println(err)
			return err
		}

		log.Printf("%s commented here %s", *comment.User.Login, *event.Comment.HTMLURL)

		// TODO(kwk): Check for event.Comment.AuthorAssociation == FIRST_TIME_CONTRIBUTOR
		// // Possible values are "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "OWNER", or "NONE".

		// tag::get_pr[]
		commentUser := *event.Comment.User.Login
		repoOwner := *event.Repo.Owner.Login
		repoName := *event.Repo.Name
		prNumber := *event.Issue.Number
		pr, _, err := gh.PullRequests.Get(context.Background(), repoOwner, repoName, prNumber)
		// end::get_pr[]
		if err != nil {
			return fmt.Errorf("failed to get pull request: %w", err)
		}

		// tag::check_mergable[]
		if !pr.GetMergeable() {
			// end::check_mergable[]
			_, _, err1 := gh.Issues.CreateComment(context.Background(), repoOwner, repoName, prNumber, &github.IssueComment{
				Body: github.String(thankYouComment + "Sorry, but this pull request is currently not mergable."),
			})
			if err1 != nil {
				err1 = fmt.Errorf("failed to write comment aobut mergability: %w", err1)
				return fmt.Errorf("pr is not mergable: %w", err1)
			}
			// TODO(kwk): Do we just want to return?
			return fmt.Errorf("pr is not mergable")
			// tag::check_mergable[]
		}
		// end::check_mergable[]

		// If there already is a check run for the same HEAD and the force
		// option is no, then say: Sorry, no can't do.
		// ----
		checkRuns, err := GetAllCheckRunsForPullRequest(gh, appInstallationID, pr)
		if err != nil {
			return fmt.Errorf("failed to get all check runs for pull request: %w", err)
		}
		currentCheckRunName := cmd.ToGithubCheckNameString()
		for _, checkRun := range checkRuns {
			if *checkRun.Name != currentCheckRunName {
				continue
			}
			// The requested check has already been run for the given PR. Let's see if a build is forced this time.
			if cmd.Force {
				// Break because we will continue with the build as planned
				break
			}

			msg := fmt.Sprintf(thankYouComment+`
The same build request exists for this pull request's SHA (%s) <a href="%s">here</a>.
Consider specifying the <code>%s=true</code> option to enforce a new build.
			`, *pr.Head.SHA, *checkRun.HTMLURL, command.CommandOptionForce)
			_, _, err := gh.Issues.CreateComment(context.Background(), repoOwner, repoName, prNumber, &github.IssueComment{
				Body: github.String(msg),
			})
			if err != nil {
				return fmt.Errorf("failed to create comment: %w", err)
			}
			// We return here because the force option was false
			return nil
		}

		// ----

		buildLogCommentID := int64(0)
		// tag::thank_you[]
		newComment, _, err := gh.Issues.CreateComment(context.Background(), repoOwner, repoName, prNumber, &github.IssueComment{
			Body: github.String(thankYouComment +
				`<sub>This very comment will be used to continously log build state changes for your request. We decided to do this in addition to using Github's Check Runs below so you can inspect previous check runs better.</sub>`,
			),
		})
		// end::thank_you[]
		if err != nil {
			return fmt.Errorf("failed to create build-log comment: %w", err)
		}
		if newComment != nil {
			buildLogCommentID = newComment.GetID()
		}

		// IDEA: We could set up one try-builder for all jobs and have that
		// try-builder trigger other jobs depending on the given properties. The
		// try builder would need to have a Trigger build step.
		// (http://docs.buildbot.net/current/manual/configuration/steps/trigger.html)
		//
		// Or we could have one try-builder fast workers and one try-builder for
		// slowers non-mandatory workers.

		//---------------------------------------------------------------------
		// Lets add a check for the try bot run
		// NOTE: It is important to first create the check so that we can pass
		//       its ID to buildbot. This way whenever buildbot tells us
		//       anything about a build we know how to reflect this in the
		//       check run on github.
		//---------------------------------------------------------------------
		opts := github.CreateCheckRunOptions{
			Name:    cmd.ToGithubCheckNameString(),
			HeadSHA: *pr.Head.SHA,
			Status:  github.String(string(CheckRunStateQueued)),
			Output: &github.CheckRunOutput{
				Title:   github.String("Buildbot Status Log"),
				Summary: github.String(WrapMsgWithTimePrefix("We're about to forward your request to buildbot.", time.Now())),
				Text:    github.String("Please wait for the URL to your buildbot job to appear here."),
				Images:  nil,
			},
		}
		optActions := []*github.CheckRunAction{
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
		}
		opts.Actions = optActions
		checkRunTryBot, _, err := gh.Checks.CreateCheckRun(context.Background(), repoOwner, repoName, opts)
		if err != nil {
			return fmt.Errorf("failed to create try bot check run: %w", err)
		}

		// To simulate latency
		// log.Printf("Sleep for 10 seconds before sending request to buildbot")
		// time.Sleep(10 * time.Second)

		// Make a buildbot try call with an empty diff (!!!)
		props := NewGithubPullRequest(pr).ToTryBotPropertyArray()
		props = append(props, fmt.Sprintf("--property=github_check_run_id=%d", *checkRunTryBot.ID))
		props = append(props, fmt.Sprintf("--property=github_app_installation_id=%d", appInstallationID))
		props = append(props, fmt.Sprintf("--property=github_build_log_comment_id=%d", buildLogCommentID))
		props = append(props, cmd.ToTryBotPropertyArray()...)
		combinedOutput, err := srv.RunTryBot(commentUser, repoOwner, repoName, props...)
		if err != nil {
			return fmt.Errorf("failed to run trybot: %s: %w", combinedOutput, err)
		}
		log.Printf("trybot command executed: %s", combinedOutput)

		return nil
	}
}

func GetAllCheckRunsForPullRequest(gh *github.Client, appInstallID int64, pr *github.PullRequest) ([]*github.CheckRun, error) {
	if gh == nil {
		return nil, fmt.Errorf("github client object is nil")
	}
	// For pagination see: https://docs.github.com/en/rest/guides/using-pagination-in-the-rest-api?apiVersion=2022-11-28
	if pr == nil {
		return nil, fmt.Errorf("pull request object is nil")
	}
	pagesRemaining := true
	pages := []*github.CheckRun{}
	listOpts := &github.ListCheckRunsOptions{
		AppID: &appInstallID,
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 30,
		},
	}
	for pagesRemaining {
		checkRunResults, resp, err := gh.Checks.ListCheckRunsForRef(context.Background(), *pr.Base.Repo.Owner.Login, *pr.Base.Repo.Name, *pr.Head.SHA, listOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to list check runs: %w", err)
		}
		if resp.NextPage == 0 {
			pagesRemaining = false
		} else {
			listOpts.ListOptions.Page = resp.NextPage
		}
		pages = append(pages, checkRunResults.CheckRuns...)
	}
	return pages, nil
}
