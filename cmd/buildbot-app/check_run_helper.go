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

func defaultCreateCheckRunOptions(pr *github.PullRequest, name string, status CheckRunState, title string, summary string, text string) github.CreateCheckRunOptions {
	return github.CreateCheckRunOptions{
		Name:    name,
		HeadSHA: *pr.Head.SHA,
		Status:  github.String(string(status)),
		Output: &github.CheckRunOutput{
			Title:   github.String(title),
			Summary: github.String(summary),
			Text:    github.String(text),
			Images: []*github.CheckRunImage{
				{
					Alt:      github.String("Buildbot App Logo"),
					ImageURL: github.String("https://raw.githubusercontent.com/kwk/buildbot-app/main/logo/logo-round-smaller.png"),
					// Caption:  github.String("Buildbot App Logo"),
				},
			},
		},
	}
}
