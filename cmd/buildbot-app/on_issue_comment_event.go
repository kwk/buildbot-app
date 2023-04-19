package main

import (
	"context"
	"fmt"
	"log"
	"regexp"

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
		if !regexp.MustCompile("^/buildbot$").MatchString(comment.GetBody()) {
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

		commentOnIssue := func(commentBody string) {
			_, _, err = gh.Issues.CreateComment(context.Background(),
				ownerLogin,
				repoName,
				prNumber,
				&github.IssueComment{
					Body: github.String(commentBody),
				})
			if err != nil {
				log.Printf("failed to create comment: %+v", err)
			}
		}

		if !pr.GetMergeable() {
			commentOnIssue("Sorry, but this PR is currently not mergable.")
			return fmt.Errorf("pr is not mergable: %w", err)
		}

		commentOnIssue(fmt.Sprintf("Thank you for using `/buildbot` [here](%s)", *event.Comment.HTMLURL))
		// gh.Issues.EditComment(context.Background(), ownerLogin, repoName, *event.Comment.ID, &github.IssueComment{
		// 	Reactions: comment.GetReactions(),
		// })

		// _, _, err = gh.Issues.EditComment(context.Background(),
		// 	ownerLogin,
		// 	repoName,
		// 	int64(prNumber),
		// 	&github.IssueComment{
		// 		Reactions: &github.Reactions{Eyes: github.Int(1)},
		// 	})
		// if err != nil {
		// 	err = fmt.Errorf("error creating reaction: %w", err)
		// 	log.Println(err)
		// 	return err
		// }

		// TODO(kwk): Add a reaction to the triggering /buildbot comment
		// 403 resource not accessible by integration
		// _, _, err = gh.Reactions.CreateCommentReaction(context.Background(), ownerLogin, repoName, *event.Comment.ID, "eyes")
		// if err != nil {
		// 	err = fmt.Errorf("error creating reaction: %w", err)
		// 	log.Println(err)
		// 	return err
		// }

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
		opts := defaultCreateCheckRunOptions(
			pr,
			fmt.Sprintf("%s's buildbot check", commentUser),
			CheckRunStateQueued,
			"Buildbot job",
			"We're starting this",
			"Nothing yet",
		)
		optActions := []*github.CheckRunAction{
			{
				Label:       "Make check required",
				Description: "Make check required to pass",
				Identifier:  "MakeMandatory",
			},
			{
				Label:       "Make check optional",
				Description: "Make check required to pass",
				Identifier:  "MakeOptional",
			},
		}
		opts.Actions = optActions
		checkRunTryBot, _, err := gh.Checks.CreateCheckRun(context.Background(), ownerLogin, repoName, opts)
		if err != nil {
			return fmt.Errorf("failed to create try bot check run: %w", err)
		}

		// Make a buildbot try call with an empty diff (!!!)
		props := NewGithubPullRequest(pr).ToTryBotPropertyArray()
		props = append(props, fmt.Sprintf("--property=github_check_run_id=%d", *checkRunTryBot.ID))
		props = append(props, fmt.Sprintf("--property=github_app_installation_id=%d", appInstallationID))
		combinedOutput, err := srv.RunTryBot(commentUser, ownerLogin, repoName, props...)
		if err != nil {
			return fmt.Errorf("failed to run trybot: %s: %w", combinedOutput, err)
		}
		log.Printf("trybot command executed: %s", combinedOutput)
		_, _, err = gh.Checks.UpdateCheckRun(context.Background(), ownerLogin, repoName, *checkRunTryBot.ID, github.UpdateCheckRunOptions{
			Name:    *checkRunTryBot.Name,
			Status:  github.String(string(CheckRunStateInProgress)),
			Actions: optActions,
		})
		if err != nil {
			return fmt.Errorf("failed to update try bot check run: %w", err)
		}

		//---------------------------------------------------------------------
		// Let's create a check run that always succeeds
		//---------------------------------------------------------------------
		opts = defaultCreateCheckRunOptions(pr,
			"Always succeed run",
			CheckRunStateQueued,
			"Placeholder check",
			"This check always succeeds",
			"You can put all kinds of `markdown` *in* [here](example.com).",
		)
		opts.DetailsURL = nil // Does this turn off the URL?
		checkRunAlwaysSucceed, _, err := gh.Checks.CreateCheckRun(context.Background(), ownerLogin, repoName, opts)
		if err != nil {
			return fmt.Errorf("failed to create always succeeding check run: %w", err)
		}
		_, _, err = gh.Checks.UpdateCheckRun(context.Background(), ownerLogin, repoName, *checkRunAlwaysSucceed.ID, github.UpdateCheckRunOptions{
			Name:       *checkRunAlwaysSucceed.Name,
			Status:     github.String(string(CheckRunStateCompleted)),
			Conclusion: github.String(string(CheckRunConclusionSuccess)),
		})
		if err != nil {
			return fmt.Errorf("failed to update always succeeding  check run: %w", err)
		}

		//---------------------------------------------------------------------
		// Let's create a check run that has some action on it (not wired back
		// atm.)
		//---------------------------------------------------------------------
		opts = defaultCreateCheckRunOptions(pr, "Action check run name", CheckRunStateQueued, "Action check run summary", "This check always succeeds", "")
		opts.Actions = []*github.CheckRunAction{
			{
				Label:       "My Button Label",
				Description: "Curious what will happen?",
				Identifier:  "OurRefForThisAction",
			},
		}
		_, _, err = gh.Checks.CreateCheckRun(context.Background(), ownerLogin, repoName, opts)
		if err != nil {
			return fmt.Errorf("failed to create check run with an action on it: %w", err)
		}

		return nil
	}
}
