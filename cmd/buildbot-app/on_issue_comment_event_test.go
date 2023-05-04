package main

import (
	"net/http"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/require"
)

// MockServer implements Server
type MockServer struct {
	mockOptions []mock.MockBackendOption
}

// NewMockServer returns a new MockServer object with the given options
func NewMockServer(options ...mock.MockBackendOption) *MockServer {
	return &MockServer{
		mockOptions: options,
	}
}

func (srv MockServer) NewGithubClient(appInstallationID int64) (*github.Client, error) {
	mockedHTTPClient := mock.NewMockedHTTPClient(srv.mockOptions...)
	return github.NewClient(mockedHTTPClient), nil
}
func (srv MockServer) RunTryBot(responsibleGithubLogin string, githubRepoOwner string, githubRepoName string, properties ...string) (string, error) {
	return "", nil
}

func issueCommentEventOK() *github.IssueCommentEvent {
	return &github.IssueCommentEvent{
		Action: github.String("created"),
		Issue: &github.Issue{
			Number:           github.Int(123),
			PullRequestLinks: &github.PullRequestLinks{},
		},
		Installation: &github.Installation{
			ID: github.Int64(1234),
		},
		Repo: &github.Repository{
			Owner: &github.User{
				Login: github.String("janedoe"),
			},
			Name: github.String("examplerepo"),
		},
		Comment: &github.IssueComment{
			Body: github.String("/buildbot"),
			User: &github.User{
				Login: github.String("johndoe"),
			},
			HTMLURL: github.String("http://github.com/jane/doe/example"),
		},
	}
}

func prOK() github.PullRequest {
	return github.PullRequest{
		Mergeable: github.Bool(true),
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{
				Owner: &github.User{
					Login: github.String("janedoe"),
				},
				Name: github.String("examplerepo"),
			},
		},
		Head: &github.PullRequestBranch{
			SHA: github.String("5da7cf6468aabc181b3c7c662539cd3e70526c1b"),
		},
	}
}

// tag::test_pr_not_mergable[]
func TestOnIssueCommentEventAny(t *testing.T) {
	// end::test_pr_not_mergable[]

	t.Run("entry level testing", func(t *testing.T) {
		t.Run("event nil", func(t *testing.T) {
			fn := OnIssueCommentEventAny(NewMockServer())
			err := fn("1234", "created", nil)
			require.NoError(t, err)
		})
		t.Run("issue is not a pull request", func(t *testing.T) {
			fn := OnIssueCommentEventAny(NewMockServer())
			event := issueCommentEventOK()
			event.Issue.PullRequestLinks = nil
			err := fn("1234", "created", event)
			require.NoError(t, err)
		})
		t.Run("deleted event", func(t *testing.T) {
			fn := OnIssueCommentEventAny(NewMockServer())
			event := issueCommentEventOK()
			event.Action = github.String("deleted")
			err := fn("1234", "deleted", event)
			require.NoError(t, err)
		})
		t.Run("comment nil", func(t *testing.T) {
			fn := OnIssueCommentEventAny(NewMockServer())
			event := issueCommentEventOK()
			event.Comment = nil
			err := fn("1234", "created", event)
			require.NoError(t, err)
		})
		t.Run("comment body nil", func(t *testing.T) {
			fn := OnIssueCommentEventAny(NewMockServer())
			event := issueCommentEventOK()
			event.Comment.Body = nil
			err := fn("1234", "created", event)
			require.NoError(t, err)
		})
		t.Run("comment body is not a command", func(t *testing.T) {
			fn := OnIssueCommentEventAny(NewMockServer())
			event := issueCommentEventOK()
			event.Comment.Body = github.String("I am a comment body and certainly not a command.")
			err := fn("1234", "created", event)
			require.NoError(t, err)
		})
	})

	t.Run("command from string", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			fn := OnIssueCommentEventAny(NewMockServer())
			event := issueCommentEventOK()
			event.Comment.Body = github.String("I am a comment body and certainly not a command.")
			err := fn("1234", "created", event)
			require.NoError(t, err)
		})
	})

	// tag::test_pr_not_mergable[]
	t.Run("pr not mergable", func(t *testing.T) {
		t.Run("comment writable", func(t *testing.T) {
			prNotMergable := prOK()
			prNotMergable.Mergeable = github.Bool(false)
			srv := NewMockServer(
				// Get PR for comment event
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					prNotMergable,
				),
				// Create comment on about PR not being mergable
				mock.WithRequestMatch(
					mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
					github.IssueComment{
						Body: github.String("blabla"),
					},
				),
			)
			fn := OnIssueCommentEventAny(srv)
			err := fn("1234", "created", issueCommentEventOK())
			require.ErrorContains(t, err, "pr is not mergable", "expected and error because pr is not mergable, yet")
		})
		// end::test_pr_not_mergable[]
		t.Run("comment not writable", func(t *testing.T) {
			prNotMergable := prOK()
			prNotMergable.Mergeable = github.Bool(false)
			srv := NewMockServer(
				// Get PR for comment event
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					prNotMergable,
				),
				// Create comment on about PR not being mergable
				mock.WithRequestMatchHandler(
					mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mock.WriteError(w, http.StatusInternalServerError, "failed to create comment")
					}),
				),
			)
			fn := OnIssueCommentEventAny(srv)
			err := fn("1234", "created", issueCommentEventOK())
			require.ErrorContains(t, err, "failed to write comment aobut mergability", "expected and error because pr is not mergable, yet")
		})
		// tag::test_pr_not_mergable[]
	})
	// end::test_pr_not_mergable[]
	// t.Run("ok", func(t *testing.T) {
	// 	pr := prOK()
	// 	srv := NewMockServer(
	// 		mock.WithRequestMatch(
	// 			mock.GetReposPullsByOwnerByRepoByPullNumber,
	// 			pr,
	// 		),
	// 		mock.WithRequestMatchPages(
	// 			mock.GetReposCommitsCheckRunsByOwnerByRepoByRef,
	// 			// github.ListCheckRunsResults {
	// 			// 	Total: github.Int(0),
	// 			// 	CheckRuns: nil,
	// 			// },
	// 			[]github.CheckRun{
	// 				{
	// 					Name: github.String("some other check run with a different name"),
	// 				},
	// 			},
	// 		),
	// 		mock.WithRequestMatchHandler(
	// 			mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
	// 			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 				mock.WriteError(
	// 					w,
	// 					http.StatusInternalServerError,
	// 					"failed to create comment",
	// 				)
	// 			}),
	// 		),
	// 	)
	// 	fn := OnIssueCommentEventAny(srv)
	// 	err := fn("1234", "created", issueCommentEventOK())
	// 	fmt.Println(err)
	// 	require.NoError(t, err, "expected and error because pr is not mergable, yet")
	// })
	// tag::test_pr_not_mergable[]
}

// end::test_pr_not_mergable[]
