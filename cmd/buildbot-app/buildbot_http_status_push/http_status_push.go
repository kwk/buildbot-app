package buildbot_http_status_push

// These data structures have been generated from an example response using:
// https://mholt.github.io/json-to-go/
//
// TODO(kwk): There's definitively room for optimization here but for now its a
// good way of showing that we can use the the HTTPStatusPush reporter from
// Buildbot:
// https://docs.buildbot.net/latest/manual/configuration/reporters/http_status.html

type Properties struct {
	GithubPullRequestNumber    []string `json:"github_pull_request_number,omitempty"`
	GithubPullRequestRepoName  []string `json:"github_pull_request_repo_name,omitempty"`
	GithubPullRequestRepoOwner []string `json:"github_pull_request_repo_owner,omitempty"`
	GithubPullRequestBaseRef   []string `json:"github_pull_request_base_ref,omitempty"`
	GithubPullRequestBaseSha   []string `json:"github_pull_request_base_sha,omitempty"`
	GithubPullRequestHeadRef   []string `json:"github_pull_request_head_ref,omitempty"`
	GithubPullRequestHeadSha   []string `json:"github_pull_request_head_sha,omitempty"`
	GithubCheckRunID           []string `json:"github_check_run_id,omitempty"`
	GithubAppInstallationID    []string `json:"github_app_installation_id,omitempty"`
	Scheduler                  []string `json:"scheduler,omitempty"`
	Buildername                []string `json:"buildername,omitempty"`
	Os                         []string `json:"os,omitempty"`
	Arch                       []string `json:"arch,omitempty"`
	OsDistro                   []string `json:"os-distro,omitempty"`
	OsVer                      []string `json:"os-ver,omitempty"`
	Workername                 []string `json:"workername,omitempty"`
	Buildnumber                []any    `json:"buildnumber,omitempty"`
	Branch                     []any    `json:"branch,omitempty"`
	Revision                   []any    `json:"revision,omitempty"`
	Repository                 []string `json:"repository,omitempty"`
	Codebase                   []string `json:"codebase,omitempty"`
	Project                    []string `json:"project,omitempty"`
	Builddir                   []string `json:"builddir,omitempty"`
}
type Buildrequest struct {
	Buildrequestid    int  `json:"buildrequestid,omitempty"`
	Buildsetid        int  `json:"buildsetid,omitempty"`
	Builderid         int  `json:"builderid,omitempty"`
	Priority          int  `json:"priority,omitempty"`
	Claimed           bool `json:"claimed,omitempty"`
	ClaimedAt         int  `json:"claimed_at,omitempty"`
	ClaimedByMasterid int  `json:"claimed_by_masterid,omitempty"`
	Complete          bool `json:"complete,omitempty"`
	Results           int  `json:"results,omitempty"`
	SubmittedAt       int  `json:"submitted_at,omitempty"`
	CompleteAt        any  `json:"complete_at,omitempty"`
	WaitedFor         bool `json:"waited_for,omitempty"`
	Properties        any  `json:"properties,omitempty"`
}
type Sourcestamps struct {
	Ssid       int    `json:"ssid,omitempty"`
	Branch     any    `json:"branch,omitempty"`
	Revision   any    `json:"revision,omitempty"`
	Project    string `json:"project,omitempty"`
	Repository string `json:"repository,omitempty"`
	Codebase   string `json:"codebase,omitempty"`
	CreatedAt  int    `json:"created_at,omitempty"`
	Patch      any    `json:"patch,omitempty"`
}
type Buildset struct {
	ExternalIdstring   any            `json:"external_idstring,omitempty"`
	Reason             string         `json:"reason,omitempty"`
	SubmittedAt        int            `json:"submitted_at,omitempty"`
	Complete           bool           `json:"complete,omitempty"`
	CompleteAt         any            `json:"complete_at,omitempty"`
	Results            int            `json:"results,omitempty"`
	Bsid               int            `json:"bsid,omitempty"`
	Sourcestamps       []Sourcestamps `json:"sourcestamps,omitempty"`
	ParentBuildid      any            `json:"parent_buildid,omitempty"`
	ParentRelationship any            `json:"parent_relationship,omitempty"`
}
type Builder struct {
	Builderid   int    `json:"builderid,omitempty"`
	Name        string `json:"name,omitempty"`
	Masterids   []int  `json:"masterids,omitempty"`
	Description any    `json:"description,omitempty"`
	Tags        []any  `json:"tags,omitempty"`
}
type Data struct {
	Buildid        int          `json:"buildid,omitempty"`
	Number         int          `json:"number,omitempty"`
	Builderid      int          `json:"builderid,omitempty"`
	Buildrequestid int          `json:"buildrequestid,omitempty"`
	Workerid       int          `json:"workerid,omitempty"`
	Masterid       int          `json:"masterid,omitempty"`
	StartedAt      int          `json:"started_at,omitempty"`
	CompleteAt     int          `json:"complete_at,omitempty"`
	Complete       bool         `json:"complete,omitempty"`
	StateString    string       `json:"state_string,omitempty"`
	Results        int          `json:"results,omitempty"`
	Properties     Properties   `json:"properties,omitempty"`
	Buildrequest   Buildrequest `json:"buildrequest,omitempty"`
	Buildset       Buildset     `json:"buildset,omitempty"`
	Parentbuild    any          `json:"parentbuild,omitempty"`
	Parentbuilder  any          `json:"parentbuilder,omitempty"`
	Builder        Builder      `json:"builder,omitempty"`
	URL            string       `json:"url,omitempty"`
}
