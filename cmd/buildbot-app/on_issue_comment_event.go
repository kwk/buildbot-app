package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v50/github"
)

func (srv *AppServer) OnIssueCommentEventAny() githubevents.IssueCommentEventHandleFunc {
	return func(deliveryID string, eventName string, event *github.IssueCommentEvent) error {
		if event == nil {
			return nil
		}
		// Only run on PR comments and not on issue comments
		if !event.Issue.IsPullRequest() {
			return nil
		}
		// Only handle new comments or edited ones
		if event.GetAction() != "edited" && event.GetAction() != "created" {
			return nil
		}
		// Check if the /buildbot command is used in the comment's body
		comment := event.GetComment()
		if comment == nil {
			return nil
		}
		commentBody := comment.GetBody()
		if !regexp.MustCompile("^/buildbot$").MatchString(commentBody) {
			return nil
		}
		log.Printf("/buildbot was used")

		// Create a github client based for this app's installation
		appInstallationID := *event.GetInstallation().ID
		gh, err := srv.NewGithubClient(appInstallationID)
		if err != nil {
			err = fmt.Errorf("error creating github client: %w", err)
			log.Println(err)
			return err
		}

		log.Printf("%s commented here %s", comment.GetUser().GetLogin(), *event.Comment.HTMLURL)

		// TODO(kwk): Check for event.Comment.AuthorAssociation == FIRST_TIME_CONTRIBUTOR
		// // Possible values are "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "OWNER", or "NONE".

		commentUser := *event.Comment.User.Login
		ownerLogin := *event.Repo.Owner.Login
		repoName := *event.Repo.Name
		prNumber := *event.Issue.Number
		pr, _, err := gh.PullRequests.Get(context.Background(), ownerLogin, repoName, prNumber)
		if err != nil {
			log.Printf("failed to get pull request: %+v", err)
		}

		commentOnIssue := func(commentBody string) *github.IssueComment {
			issueComment, _, err := gh.Issues.CreateComment(context.Background(),
				ownerLogin,
				repoName,
				prNumber,
				&github.IssueComment{
					Body: github.String(commentBody),
				})
			if err != nil {
				log.Printf("failed to create comment: %+v", err)
				return nil
			}
			return issueComment
		}

		if !pr.GetMergeable() {
			commentOnIssue("Sorry, but this PR is currently not mergable.")
			return fmt.Errorf("pr is not mergable: %w", err)
		}

		buildLogCommentID := int64(0)
		newComment := commentOnIssue(fmt.Sprintf(`Thank you @%s for using the <a href="todo:link-to-documentation-here"><code>/buildbot</code></a> command <a href="%s">here</a>!
<sub>This very comment will be used to continously log build state changes for your request. We decided to do this in addition to using Github's Check Runs below so you can inspect previous check runs better.</sub>
`, *event.Comment.User.Login, *event.Comment.HTMLURL))
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
			Name:    fmt.Sprintf("@%s's buildbot check: %s", commentUser, commentBody),
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
		checkRunTryBot, _, err := gh.Checks.CreateCheckRun(context.Background(), ownerLogin, repoName, opts)
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
		combinedOutput, err := srv.RunTryBot(commentUser, ownerLogin, repoName, props...)
		if err != nil {
			return fmt.Errorf("failed to run trybot: %s: %w", combinedOutput, err)
		}
		log.Printf("trybot command executed: %s", combinedOutput)

		return nil
	}
}
