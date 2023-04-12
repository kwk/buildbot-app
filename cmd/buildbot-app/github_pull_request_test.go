package main

import (
	"reflect"
	"testing"
)

func TestGithubPullRequest_ToTryBotPropertyArray(t *testing.T) {
	type fields struct {
		Number        int
		BaseRepoName  string
		BaseRepoOwner string
		BaseRef       string
		BaseSHA       string
		HeadRef       string
		HeadSHA       string
		// GithubTriggerCommentHTMLURL    string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "ok_all_fields_set",
			fields: fields{
				3,
				"base.name",
				"base.owner",
				"base.ref",
				"base.sha",
				"head.ref",
				"head.sha",
				// "",
			},
			want: []string{
				"--property=github_pull_request_number=3",
				"--property=github_pull_request_repo_name=base.name",
				"--property=github_pull_request_repo_owner=base.owner",
				"--property=github_pull_request_base_ref=base.ref",
				"--property=github_pull_request_base_sha=base.sha",
				"--property=github_pull_request_head_ref=head.ref",
				"--property=github_pull_request_head_sha=head.sha",
				// "--property=github_trigger_comment_html_url=",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &GithubPullRequest{
				Number:        tt.fields.Number,
				BaseRepoName:  tt.fields.BaseRepoName,
				BaseRepoOwner: tt.fields.BaseRepoOwner,
				BaseRef:       tt.fields.BaseRef,
				BaseSHA:       tt.fields.BaseSHA,
				HeadRef:       tt.fields.HeadRef,
				HeadSHA:       tt.fields.HeadSHA,
			}
			if got := pr.ToTryBotPropertyArray(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GithubPullRequest.ToTryBotPropertyArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
