package main

import (
	"fmt"
	"reflect"

	"github.com/google/go-github/v50/github"
)

// GithubPullRequest contains all the information we need to identify a PR. This
// is also what we sent to Buildbot to identify a PR.
type GithubPullRequest struct {
	Number        int    `json:"github_pull_request_number" binding:"required"`
	BaseRepoName  string `json:"github_pull_request_repo_name" binding:"required"`
	BaseRepoOwner string `json:"github_pull_request_repo_owner" binding:"required"`
	BaseRef       string `json:"github_pull_request_base_ref" binding:"required"`
	BaseSHA       string `json:"github_pull_request_base_sha" binding:"required"`
	HeadRef       string `json:"github_pull_request_head_ref" binding:"required"`
	HeadSHA       string `json:"github_pull_request_head_sha" binding:"required"`
	// GithubTriggerCommentHTMLURL string `json:"github_trigger_comment_html_url"`
}

// ToTryBotPropertyArray returns a string list that you can pass as properties
// to AppServer's RunTryBot(). We use the json field tag to name the properties.
func (pr *GithubPullRequest) ToTryBotPropertyArray() []string {
	v := reflect.ValueOf(*pr)
	typeOfS := v.Type()

	numFields := v.NumField()
	properties := make([]string, numFields)
	for i := 0; i < numFields; i++ {
		properties[i] = fmt.Sprintf("--property=%s=%v", typeOfS.Field(i).Tag.Get("json"), v.Field(i).Interface())
	}
	return properties
}

func NewGithubPullRequest(pr *github.PullRequest) *GithubPullRequest {
	if pr == nil {
		return nil
	}
	return &GithubPullRequest{
		Number:        pr.GetNumber(),
		BaseRepoName:  *pr.Base.Repo.Name,
		BaseRepoOwner: *pr.Base.Repo.Owner.Login,
		BaseRef:       *pr.Base.Ref,
		BaseSHA:       *pr.Base.SHA,
		HeadRef:       *pr.Head.Ref,
		HeadSHA:       *pr.Head.SHA,
	}
}
