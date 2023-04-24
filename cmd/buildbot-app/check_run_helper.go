package main

import (
	"github.com/google/go-github/v50/github"
)

type CheckRunState string

const CheckRunStateQueued CheckRunState = "queued"
const CheckRunStateInProgress CheckRunState = "in_progress"
const CheckRunStateCompleted CheckRunState = "completed"

type CheckRunConclusion string

const CheckRunConclusionSuccess CheckRunConclusion = "success"
const CheckRunConclusionFailure CheckRunConclusion = "failure"
const CheckRunConclusionNeutral CheckRunConclusion = "neutral"
const CheckRunConclusionCancelled CheckRunConclusion = "cancelled"
const CheckRunConclusionSkipped CheckRunConclusion = "skipped"
const CheckRunConclusionTimedOut CheckRunConclusion = "timed_out"
const CheckRunConclusionActionRequired CheckRunConclusion = "action_required"

// See also https://docs.buildbot.net/latest/developer/results.html#build-result-codes
func CheckRunStateFromBuildbotResult(resultCode int) CheckRunConclusion {
	switch resultCode {
	case 0: // success
		return CheckRunConclusionSuccess
	case 1: // warning
		return CheckRunConclusionNeutral
	case 2: // failure
		return CheckRunConclusionFailure
	case 3: // skipped
		return CheckRunConclusionSkipped
	case 4: // exception
		return CheckRunConclusionFailure
	case 5: // retry
		return CheckRunConclusionFailure
	case 6: // cancelled
		return CheckRunConclusionCancelled
	}
	return CheckRunConclusionFailure
}

func defaultCreateCheckRunOptions(pr *github.PullRequest, name string, status CheckRunState, title string, summary string, text string) github.CreateCheckRunOptions {
	return github.CreateCheckRunOptions{
		Name:    name,
		HeadSHA: *pr.Head.SHA,
		Status:  github.String(string(status)),
		Output: &github.CheckRunOutput{
			Title:   github.String(title),
			Summary: github.String(summary),
			Text:    github.String(text),
			Images:  nil,
			// Images: []*github.CheckRunImage{
			// 	{
			// 		Alt:      github.String("Buildbot App Logo"),
			// 		ImageURL: github.String("https://raw.githubusercontent.com/kwk/buildbot-app/main/logo/logo-round-small.png"),
			// 		// Caption:  github.String("Buildbot App Logo"),
			// 	},
			// },
		},
	}
}
